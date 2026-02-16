package canvas

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	// "fyne.io/fyne/v2/driver/desktop" // No longer needed
	"fyne.io/fyne/v2/widget"

	"mosugo/internal/cards"
	"mosugo/internal/tools"
)

const (
	GridSize = 40
	MinZoom  = 0.1
	MaxZoom  = 5.0
)

func snap(v float32) float32 {
	return float32(math.Round(float64(v)/GridSize)) * GridSize
}

// Convert Screen coordinates to World coordinates
func (c *MosugoCanvas) ScreenToWorld(pos fyne.Position) fyne.Position {
	// WorldX = (ScreenX - Offset.X) / Zoom
	x := float32(pos.X-c.Offset.X) / c.Zoom
	y := float32(pos.Y-c.Offset.Y) / c.Zoom
	return fyne.NewPos(x, y)
}

// Convert World coordinates to Screen coordinates
func (c *MosugoCanvas) WorldToScreen(pos fyne.Position) fyne.Position {
	// ScreenX = (WorldX * Zoom) + Offset.X
	x := float32(pos.X*c.Zoom + c.Offset.X)
	y := float32(pos.Y*c.Zoom + c.Offset.Y)
	return fyne.NewPos(x, y)
}

type MosugoCanvas struct {
	widget.BaseWidget

	// UI
	Grid    *canvas.Raster
	Content *fyne.Container // Container for cards

	// State
	Offset      fyne.Position // The camera position
	Zoom        float32
	CurrentTool tools.ToolType

	// Drag Logic (for creating new cards)
	dragStart    fyne.Position     // Screen position where drag started
	ghostRect    *canvas.Rectangle // Visual feedback for new card size
	isDragging   bool
	selectedCard *cards.MosuWidget // Currently manipulated card
	dragOffset   fyne.Position     // Offset from card top-left during move
}

func NewMosugoCanvas() *MosugoCanvas {
	c := &MosugoCanvas{
		Zoom:   1.0,
		Offset: fyne.NewPos(0, 0),
	}
	c.ExtendBaseWidget(c)

	// Lighter gray, larger dots (3x3)
	dotColor := color.RGBA{150, 150, 150, 255}
	bgColor := color.RGBA{250, 250, 250, 255}

	// Pass 'c' to make grid dynamic based on Offset/Zoom
	c.Grid = DotGridPattern(c, GridSize, dotColor, bgColor)

	// Content container holds the actual cards.
	// Objects are placed here at their "World Position" + "Offset".
	c.Content = container.NewWithoutLayout()

	// Initialize Ghost Rect (light green, thick border)
	c.ghostRect = canvas.NewRectangle(color.RGBA{144, 238, 144, 100})
	c.ghostRect.StrokeColor = color.RGBA{0, 200, 0, 255}
	c.ghostRect.StrokeWidth = 3 // Thicker for better visibility
	c.ghostRect.Hide()

	// Add GhostRect last to ensure it's on top
	c.Content.Add(c.ghostRect)

	return c
}

// CreateRenderer defines the widget layout
func (c *MosugoCanvas) CreateRenderer() fyne.WidgetRenderer {
	return &mosugoRenderer{canvas: c}
}

// Renderer implementation
type mosugoRenderer struct {
	canvas *MosugoCanvas
}

func (r *mosugoRenderer) Destroy() {}

func (r *mosugoRenderer) Layout(size fyne.Size) {
	// 1. Grid Background Layout
	// We size the grid to cover the whole widget view
	if r.canvas.Grid != nil {
		r.canvas.Grid.Resize(size)
		r.canvas.Grid.Move(fyne.NewPos(0, 0))
	}

	// 2. Content Layout (The Critical Part)
	// We update the position/size of every card manually for Zoom.
	// We no longer rely on `r.canvas.Content.Move(Offset)`.
	// Instead, the container stays at (0,0) and fills the viewport.
	// But `SetContent` allows us to manipulate *children*.

	if r.canvas.Content != nil {
		// New Strategy:
		// Keep c.Content filling the viewport
		r.canvas.Content.Resize(size)
		r.canvas.Content.Move(fyne.NewPos(0, 0))

		// Update all child positions here using World Logic
		for _, obj := range r.canvas.Content.Objects {
			if mosuW, ok := obj.(*cards.MosuWidget); ok {
				// Apply Zoom & Pan logic
				// ScreenPos = (WorldPos * Zoom) + PanOffset
				// ScreenSize = WorldSize * Zoom

				screenPos := r.canvas.WorldToScreen(mosuW.WorldPos)
				screenSize := fyne.NewSize(
					mosuW.WorldSize.Width*r.canvas.Zoom,
					mosuW.WorldSize.Height*r.canvas.Zoom,
				)

				mosuW.Move(screenPos)
				mosuW.Resize(screenSize)
			} else if rect, ok := obj.(*canvas.Rectangle); ok && rect == r.canvas.ghostRect {
				// Ensure ghost rect is drawn on top and visible
				r.canvas.Content.Objects = append(removeObj(r.canvas.Content.Objects, rect), rect)
			}
		}
	}
}

func removeObj(s []fyne.CanvasObject, r fyne.CanvasObject) []fyne.CanvasObject {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func (r *mosugoRenderer) MinSize() fyne.Size {
	return fyne.NewSize(100, 100)
}

func (r *mosugoRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.canvas.Grid, r.canvas.Content}
}

func (r *mosugoRenderer) Refresh() {
	r.Layout(r.canvas.Size())
	canvas.Refresh(r.canvas.Content)
}

// --- Interaction Logic ---

func (c *MosugoCanvas) Tapped(e *fyne.PointEvent) {
	// Optional: Handle selection via click instead of drag-start
	// For now, Dragged handles both "grab" and "move"
}

// Dragged implements fyne.Draggable
func (c *MosugoCanvas) Dragged(e *fyne.DragEvent) {
	switch c.CurrentTool {
	case tools.ToolPan:
		// Move the "Camera" (Offset)
		// Note: Dragged gives delta.
		c.Offset = c.Offset.Add(e.Dragged)
		c.Content.Refresh() // Force redraw of all cards with new offset

	case tools.ToolSelect:
		// Logic to move existing cards
		if !c.isDragging {
			// Find card under mouse
			found := false

			// Iterate reverse to find top-most
			for i := len(c.Content.Objects) - 1; i >= 0; i-- {
				obj := c.Content.Objects[i]
				if card, ok := obj.(*cards.MosuWidget); ok {
					// Check bounds (Screen Coords)
					// But wait, obj.Position() is relative to c.Content (which is (0,0)).
					// So obj.Position() is effectively Screen Pos.

					// However, e.Position is relative to MosugoCanvas (0,0).
					// c.Content is at (0,0). So coordinates match.

					p := card.Position()
					s := card.Size()
					if e.Position.X >= p.X && e.Position.X <= p.X+s.Width &&
						e.Position.Y >= p.Y && e.Position.Y <= p.Y+s.Height {

						c.selectedCard = card
						c.dragOffset = e.Position.Subtract(p) // Offset from top-left
						c.isDragging = true
						found = true
						break
					}
				}
			}
			if !found {
				return // Dragging on empty space does nothing (or could pan?)
			}
		}

		if c.isDragging && c.selectedCard != nil {
			// Calculate new Top-Left (Screen)
			newScreenPos := e.Position.Subtract(c.dragOffset)

			// Convert to World
			// World = (Screen - Offset) / Zoom
			rawWorldX := (float32(newScreenPos.X) - c.Offset.X) / c.Zoom
			rawWorldY := (float32(newScreenPos.Y) - c.Offset.Y) / c.Zoom

			// Build Snap
			snapX := snap(rawWorldX)
			snapY := snap(rawWorldY)

			// Update Card World Pos
			c.selectedCard.WorldPos = fyne.NewPos(snapX, snapY)

			// Force Refresh to update layout
			c.Content.Refresh()
		}

	case tools.ToolCard:
		// We need logic to start a drag vs continue a drag.
		// Fyne's Dragged() is continuous.
		// We approximate DragStart by checking a flag or checking if ghost is hidden.

		if !c.isDragging {
			c.isDragging = true
			// Calculate Start Screen Position
			c.dragStart = e.Position.Subtract(e.Dragged)

			// Show visual feedback
			c.ghostRect.Show()
		}

		// Calculate Current Mouse Pos relative to widget
		currPos := e.Position

		// Determine geometry (Left/Top/Width/Height)
		x1 := math.Min(float64(c.dragStart.X), float64(currPos.X))
		y1 := math.Min(float64(c.dragStart.Y), float64(currPos.Y))
		x2 := math.Max(float64(c.dragStart.X), float64(currPos.X))
		y2 := math.Max(float64(c.dragStart.Y), float64(currPos.Y))

		// Ghost Box should mimic final world coordinates snapping, but display in SCREEN coords

		// 1. Convert Screen Rect -> Raw World Rect
		// WorldX1 = (x1 - OffsetX) / Zoom
		rawWorldX1 := (float32(x1) - c.Offset.X) / c.Zoom
		rawWorldY1 := (float32(y1) - c.Offset.Y) / c.Zoom

		rawWorldX2 := (float32(x2) - c.Offset.X) / c.Zoom
		rawWorldY2 := (float32(y2) - c.Offset.Y) / c.Zoom

		// 2. Snap World Coordinates
		snapX1 := snap(rawWorldX1)
		snapY1 := snap(rawWorldY1)
		snapW := snap(rawWorldX2 - rawWorldX1)
		snapH := snap(rawWorldY2 - rawWorldY1)

		// 3. Convert Snapped World -> Screen Rect to display ghost
		// ScreenX = (SnapX1 * Zoom) + OffsetX
		screenGhostX := (snapX1 * c.Zoom) + c.Offset.X
		screenGhostY := (snapY1 * c.Zoom) + c.Offset.Y
		screenGhostW := snapW * c.Zoom
		screenGhostH := snapH * c.Zoom

		c.ghostRect.Move(fyne.NewPos(screenGhostX, screenGhostY))
		c.ghostRect.Resize(fyne.NewSize(screenGhostW, screenGhostH))
		c.Content.Refresh() // Ensure ghost is redrawn
	}
}

// DragEnd implements fyne.Draggable
func (c *MosugoCanvas) DragEnd() {
	if c.CurrentTool == tools.ToolCard && c.isDragging {
		// Finalize Creation
		// Don't reset isDragging yet, use it to guard logic

		rectSize := c.ghostRect.Size()
		rectPos := c.ghostRect.Position()

		// Hide ghost first
		c.ghostRect.Hide()

		// Reset Valid Drag State
		c.isDragging = false

		// Because ghostRect is now in SCREEN coordinates,
		// we must convert back to WORLD coordinates to store the card properly.
		// WorldX = (ScreenX - OffsetX) / Zoom
		// WorldSize = ScreenSize / Zoom

		worldW := rectSize.Width / c.Zoom
		worldH := rectSize.Height / c.Zoom

		worldX := (rectPos.X - c.Offset.X) / c.Zoom
		worldY := (rectPos.Y - c.Offset.Y) / c.Zoom

		// Minimum size check (GridSize is 40, so let's say >= 40)
		if worldW < GridSize || worldH < GridSize {
			c.Content.Refresh()
			return
		}

		// Create actual card
		var defaultCardColor color.Color = color.White

		cardID := fmt.Sprintf("card_%d", len(c.Content.Objects))
		newCard := cards.NewMosuWidget(cardID, defaultCardColor)

		// Store WORLD coordinates
		newCard.WorldPos = fyne.NewPos(worldX, worldY)
		newCard.WorldSize = fyne.NewSize(worldW, worldH)

		// Set initial Screen Size/Pos (Layout will update it later too)
		screenPos := c.WorldToScreen(newCard.WorldPos)
		screenSize := fyne.NewSize(
			newCard.WorldSize.Width*c.Zoom,
			newCard.WorldSize.Height*c.Zoom,
		)
		newCard.Move(screenPos)
		newCard.Resize(screenSize)
		c.Content.Add(newCard)
		c.Content.Refresh() // Refresh container specifically
	}
}

func (c *MosugoCanvas) ApplyZoom(delta float32) {
	c.Zoom += delta
	if c.Zoom < MinZoom {
		c.Zoom = MinZoom
	}
	if c.Zoom > MaxZoom {
		c.Zoom = MaxZoom
	}
	c.Refresh()
}
