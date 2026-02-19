package tools

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"

	"mosugo/internal/cards"
	"mosugo/internal/theme"
)

// Canvas interface allows tools to interact with the main MosugoCanvas.
type Canvas interface {
	Refresh()

	// Coordinate Helpers
	ScreenToWorld(pos fyne.Position) fyne.Position
	WorldToScreen(pos fyne.Position) fyne.Position
	Snap(v float32) float32
	SnapUp(v float32) float32

	// properties
	GetOffset() fyne.Position
	SetOffset(pos fyne.Position)
	GetScale() float32

	// Object modification
	AddObject(o fyne.CanvasObject)
	RemoveObject(o fyne.CanvasObject)
	ContentContainer() *fyne.Container
	AddStroke(p1, p2 fyne.Position)

	// Visual helpers
	GhostRect() *canvas.Rectangle

	// State needed for tools
	GetSelectedCard() *cards.MosuWidget
	SetSelectedCard(c *cards.MosuWidget)
}

type Tool interface {
	Name() string
	Cursor() desktop.Cursor
	OnTapped(c Canvas, e *fyne.PointEvent)
	OnDragged(c Canvas, e *fyne.DragEvent)
	OnDragEnd(c Canvas)
}

type SelectTool struct {
	isMovingCard bool
	rawPos       fyne.Position
}

func (t *SelectTool) Name() string           { return "Select Tool" }
func (t *SelectTool) Cursor() desktop.Cursor { return desktop.DefaultCursor }

func (t *SelectTool) OnTapped(c Canvas, e *fyne.PointEvent) {
	// Check if we tapped on a card to select it
	objects := c.ContentContainer().Objects
	// Iterate backwards to hit top-most first
	for i := len(objects) - 1; i >= 0; i-- {
		obj := objects[i]
		if mosuW, ok := obj.(*cards.MosuWidget); ok {
			screenPos := c.WorldToScreen(mosuW.WorldPos)
			size := mosuW.Size() // Screen size

			if e.Position.X >= screenPos.X && e.Position.X <= screenPos.X+size.Width &&
				e.Position.Y >= screenPos.Y && e.Position.Y <= screenPos.Y+size.Height {

				if c.GetSelectedCard() != nil && c.GetSelectedCard() != mosuW {
					c.GetSelectedCard().SetSelected(false)
				}
				c.SetSelectedCard(mosuW)
				mosuW.SetSelected(true)
				c.Refresh()
				return
			}
		}
	}

	// Deselect if clicked empty space
	if c.GetSelectedCard() != nil {
		c.GetSelectedCard().SetSelected(false)
		c.SetSelectedCard(nil)
		c.Refresh()
	}
}

func (t *SelectTool) OnDragged(c Canvas, e *fyne.DragEvent) {
	// 1. Identify where the drag started
	dragStartScreen := e.Position.Subtract(e.Dragged)

	// 2. If we aren't already moving a card, check if we clicked on one
	if !t.isMovingCard {
		scale := c.GetScale()
		if scale <= 0 {
			scale = 1.0
		}

		objects := c.ContentContainer().Objects
		// Iterate backwards (top-most first)
		for i := len(objects) - 1; i >= 0; i-- {
			obj := objects[i]
			if mosuW, ok := obj.(*cards.MosuWidget); ok {
				sPos := c.WorldToScreen(mosuW.WorldPos)
				sSize := mosuW.Size()

				// Hit Test on Drag Start Position
				if dragStartScreen.X >= sPos.X && dragStartScreen.X <= sPos.X+sSize.Width &&
					dragStartScreen.Y >= sPos.Y && dragStartScreen.Y <= sPos.Y+sSize.Height {

					// Select this card immediately
					if c.GetSelectedCard() != nil && c.GetSelectedCard() != mosuW {
						c.GetSelectedCard().SetSelected(false)
					}
					c.SetSelectedCard(mosuW)
					mosuW.SetSelected(true)
					t.isMovingCard = true // Start moving mode
					t.rawPos = mosuW.WorldPos
					break // Found the target
				}
			}
		}
	}

	// 3. Move Logic
	if t.isMovingCard && c.GetSelectedCard() != nil {
		card := c.GetSelectedCard()

		scale := c.GetScale()
		if scale <= 0 {
			scale = 1.0
		}

		// Convert screen delta to world delta
		dx := e.Dragged.DX / scale
		dy := e.Dragged.DY / scale

		t.rawPos.X += dx
		t.rawPos.Y += dy

		card.WorldPos.X = c.Snap(t.rawPos.X)
		card.WorldPos.Y = c.Snap(t.rawPos.Y)

		c.Refresh()
		return
	}

	// 4. Pan Logic (Fallback if no card hit)
	current := c.GetOffset()
	c.SetOffset(fyne.NewPos(
		current.X+e.Dragged.DX,
		current.Y+e.Dragged.DY,
	))
	c.Refresh()
}

func (t *SelectTool) OnDragEnd(c Canvas) {
	if t.isMovingCard && c.GetSelectedCard() != nil {
		card := c.GetSelectedCard()
		card.WorldPos.X = c.Snap(card.WorldPos.X)
		card.WorldPos.Y = c.Snap(card.WorldPos.Y)
		c.Refresh()
	}
	t.isMovingCard = false
}

type CardTool struct {
	startPos   fyne.Position
	isDragging bool
}

func (t *CardTool) OnTapped(c Canvas, e *fyne.PointEvent) {

}

func (t *CardTool) Name() string           { return "Card Tool" }
func (t *CardTool) Cursor() desktop.Cursor { return desktop.CrosshairCursor }

func (t *CardTool) OnDragged(c Canvas, e *fyne.DragEvent) {
	// Create New Card (Ghost Rect)
	if !t.isDragging {
		// Start Creation
		t.isDragging = true
		// Store start in World Coordinates
		t.startPos = c.ScreenToWorld(e.Position.Subtract(e.Dragged))

		// Snap the start position immediately so the anchor point is consistent
		t.startPos.X = c.Snap(t.startPos.X)
		t.startPos.Y = c.Snap(t.startPos.Y)

		c.GhostRect().Show()
	}

	// Current Mouse in World
	currWorld := c.ScreenToWorld(e.Position)

	// Calculate bounds
	x1 := float32(math.Min(float64(t.startPos.X), float64(currWorld.X)))
	y1 := float32(math.Min(float64(t.startPos.Y), float64(currWorld.Y)))
	x2 := float32(math.Max(float64(t.startPos.X), float64(currWorld.X)))
	y2 := float32(math.Max(float64(t.startPos.Y), float64(currWorld.Y)))

	// Snap Logic - round to nearest grid cell
	snapX1 := c.Snap(x1)
	snapY1 := c.Snap(y1)
	snapX2 := c.SnapUp(x2)
	snapY2 := c.SnapUp(y2)

	// Ensure minimal size of 1 grid cell
	if snapX2 <= snapX1 {
		snapX2 = c.SnapUp(snapX1 + 1)
	}
	if snapY2 <= snapY1 {
		snapY2 = c.SnapUp(snapY1 + 1)
	}

	w := snapX2 - snapX1
	h := snapY2 - snapY1

	// Update Ghost Rect
	screenPos := c.WorldToScreen(fyne.NewPos(snapX1, snapY1))
	screenSize := fyne.NewSize(w*c.GetScale(), h*c.GetScale())

	c.GhostRect().Move(screenPos)
	c.GhostRect().Resize(screenSize)
	c.Refresh()
}

func (t *CardTool) OnDragEnd(c Canvas) {
	if t.isDragging {
		t.isDragging = false
		c.GhostRect().Hide()

		// Determine final loop from ghost rect
		gPos := c.GhostRect().Position()
		gSize := c.GhostRect().Size()

		worldPos := c.ScreenToWorld(gPos)
		worldPos.X = c.Snap(worldPos.X)
		worldPos.Y = c.Snap(worldPos.Y)

		worldW := gSize.Width / c.GetScale()
		worldH := gSize.Height / c.GetScale()

		if worldW <= 1 || worldH <= 1 {
			c.Refresh()
			return
		}

		cardID := fmt.Sprintf("card_%d", len(c.ContentContainer().Objects))
		newCard := cards.NewMosuWidget(cardID, theme.CardBg)

		// This avoids any Screen -> World floating point drift
		newCard.WorldPos = fyne.NewPos(c.Snap(worldPos.X), c.Snap(worldPos.Y))
		newCard.WorldSize = fyne.NewSize(c.SnapUp(worldW), c.SnapUp(worldH))
		newCard.RefreshContent() // Ensure content is parsed (default empty)

		c.AddObject(newCard)

		// Animate
		AnimateCardBounce(c, newCard)

		c.Refresh()
	}
}

type DrawTool struct {
	lastDrawPos fyne.Position
	isDrawing   bool
}

func (t *DrawTool) Name() string           { return "Draw Tool" }
func (t *DrawTool) Cursor() desktop.Cursor { return desktop.TextCursor } // Pencil-ish

func (t *DrawTool) OnTapped(c Canvas, e *fyne.PointEvent) {}

func (t *DrawTool) OnDragged(c Canvas, e *fyne.DragEvent) {
	if !t.isDrawing {
		t.isDrawing = true
		t.lastDrawPos = e.Position
	}

	if math.Abs(float64(e.Position.X-t.lastDrawPos.X)) < 2 && math.Abs(float64(e.Position.Y-t.lastDrawPos.Y)) < 2 {
		return
	}

	startWorld := c.ScreenToWorld(t.lastDrawPos)
	endWorld := c.ScreenToWorld(e.Position)

	c.AddStroke(startWorld, endWorld)

	t.lastDrawPos = e.Position
	c.Refresh()
}

func (t *DrawTool) OnDragEnd(c Canvas) {
	t.isDrawing = false
}

// Erase Tool
type EraseTool struct{}

func (t *EraseTool) Name() string           { return "Erase Tool" }
func (t *EraseTool) Cursor() desktop.Cursor { return desktop.HResizeCursor } // Placeholder

func (t *EraseTool) OnTapped(c Canvas, e *fyne.PointEvent) {
	t.eraseAt(c, e.Position)
}
func (t *EraseTool) OnDragged(c Canvas, e *fyne.DragEvent) {
	t.eraseAt(c, e.Position)
}
func (t *EraseTool) OnDragEnd(c Canvas) {}

func (t *EraseTool) eraseAt(c Canvas, screenPos fyne.Position) {

	objects := c.ContentContainer().Objects
	for i := len(objects) - 1; i >= 0; i-- {
		obj := objects[i]

		// Handle Cards
		if mosuW, ok := obj.(*cards.MosuWidget); ok {
			sPos := c.WorldToScreen(mosuW.WorldPos)
			sPos = mosuW.Position()
			sSize := mosuW.Size()

			if screenPos.X >= sPos.X && screenPos.X <= sPos.X+sSize.Width &&
				screenPos.Y >= sPos.Y && screenPos.Y <= sPos.Y+sSize.Height {
				c.RemoveObject(obj)
				c.Refresh()
				return
			}
		}

		if line, ok := obj.(*canvas.Line); ok {

			if pointNearSegment(screenPos, line.Position1, line.Position2, 10.0) {
				c.RemoveObject(obj)
				c.Refresh()
				return
			}
		}
	}
}

func pointNearSegment(p, a, b fyne.Position, threshold float32) bool {
	x0, y0 := float64(p.X), float64(p.Y)
	x1, y1 := float64(a.X), float64(a.Y)
	x2, y2 := float64(b.X), float64(b.Y)

	dx := x2 - x1
	dy := y2 - y1
	lengthSq := dx*dx + dy*dy

	var dist float64
	if lengthSq < 0.0001 {
		dist = math.Sqrt((x0-x1)*(x0-x1) + (y0-y1)*(y0-y1))
	} else {
		t := ((x0-x1)*dx + (y0-y1)*dy) / lengthSq
		if t < 0 {
			t = 0
		} else if t > 1 {
			t = 1
		}

		projX := x1 + t*dx
		projY := y1 + t*dy
		dist = math.Sqrt((x0-projX)*(x0-projX) + (y0-projY)*(y0-projY))
	}
	return dist <= float64(threshold)
}
