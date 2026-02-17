package canvas

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	"mosugo/internal/cards"
	"mosugo/internal/theme"
	"mosugo/internal/tools"
)

const (
	GridSize = 16
)

func snap(v float32) float32 {
	return float32((int(v) / GridSize) * GridSize)
}

func snapUp(v float32) float32 {
	cells := int(v) / GridSize
	if int(v)%GridSize != 0 {
		cells++
	}
	return float32(cells * GridSize)
}

func (c *MosugoCanvas) ScreenToWorld(pos fyne.Position) fyne.Position {
	x := float32(pos.X - c.Offset.X)
	y := float32(pos.Y - c.Offset.Y)
	return fyne.NewPos(x, y)
}

func (c *MosugoCanvas) WorldToScreen(pos fyne.Position) fyne.Position {
	x := float32(pos.X + c.Offset.X)
	y := float32(pos.Y + c.Offset.Y)
	return fyne.NewPos(x, y)
}

type MosugoCanvas struct {
	widget.BaseWidget

	Grid    *canvas.Raster
	Content *fyne.Container

	Offset      fyne.Position
	CurrentTool tools.ToolType

	dragStart    fyne.Position
	ghostRect    *canvas.Rectangle
	isDragging   bool
	selectedCard *cards.MosuWidget
	dragOffset   fyne.Position

	isPanning bool
	panStart  fyne.Position

	isDraggingCard bool
	cardDragStart  fyne.Position

	isDrawing     bool
	currentStroke []*canvas.Line
	strokes       [][]*canvas.Line

	// Drawing smoothing
	lastDrawPos fyne.Position

	StrokeWidth float32
	StrokeColor color.Color
}

func NewMosugoCanvas() *MosugoCanvas {
	c := &MosugoCanvas{
		Offset:      fyne.NewPos(0, 0),
		CurrentTool: tools.ToolCard,
		StrokeWidth: 2,
		StrokeColor: theme.InkGrey,
	}
	c.ExtendBaseWidget(c)

	c.Grid = BoxGridPattern(c, GridSize, theme.GridLine, theme.GridBg)
	c.Content = container.NewWithoutLayout()

	c.ghostRect = canvas.NewRectangle(color.RGBA{100, 150, 255, 40})
	c.ghostRect.StrokeColor = theme.SelectionBlue
	c.ghostRect.StrokeWidth = 2
	c.ghostRect.Hide()

	c.Content.Add(c.ghostRect)

	return c
}

func (c *MosugoCanvas) CreateRenderer() fyne.WidgetRenderer {
	return &mosugoRenderer{canvas: c}
}

type mosugoRenderer struct {
	canvas *MosugoCanvas
}

func (r *mosugoRenderer) Destroy() {}

func (r *mosugoRenderer) Layout(size fyne.Size) {
	if r.canvas.Grid != nil {
		r.canvas.Grid.Resize(size)
		r.canvas.Grid.Move(fyne.NewPos(0, 0))
	}

	if r.canvas.Content != nil {
		r.canvas.Content.Resize(size)
		r.canvas.Content.Move(fyne.NewPos(0, 0))

		for _, obj := range r.canvas.Content.Objects {
			if mosuW, ok := obj.(*cards.MosuWidget); ok {
				worldPos := mosuW.WorldPos
				worldSize := mosuW.WorldSize

				screenX := float32(int(worldPos.X) + int(r.canvas.Offset.X))
				screenY := float32(int(worldPos.Y) + int(r.canvas.Offset.Y))

				mosuW.Move(fyne.NewPos(screenX, screenY))
				mosuW.Resize(worldSize)
			} else if rect, ok := obj.(*canvas.Rectangle); ok && rect == r.canvas.ghostRect {
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

func (c *MosugoCanvas) Cursor() desktop.Cursor {
	switch c.CurrentTool {
	case tools.ToolCard:
		return desktop.CrosshairCursor
	case tools.ToolDraw:
		return desktop.TextCursor // Closest to a 'pencil' or precision cursor in standard set
	case tools.ToolErase:
		return desktop.HResizeCursor // Placeholder, or use Default if no better option
	}
	return desktop.DefaultCursor
}

func (c *MosugoCanvas) Tapped(e *fyne.PointEvent) {
	for _, obj := range c.Content.Objects {
		if mosuW, ok := obj.(*cards.MosuWidget); ok {
			screenPos := fyne.NewPos(
				float32(int(mosuW.WorldPos.X)+int(c.Offset.X)),
				float32(int(mosuW.WorldPos.Y)+int(c.Offset.Y)),
			)
			if e.Position.X >= screenPos.X && e.Position.X <= screenPos.X+mosuW.WorldSize.Width &&
				e.Position.Y >= screenPos.Y && e.Position.Y <= screenPos.Y+mosuW.WorldSize.Height {
				if c.selectedCard != nil {
					c.selectedCard.SetSelected(false)
				}
				c.selectedCard = mosuW
				mosuW.SetSelected(true)
				c.isDraggingCard = true
				c.cardDragStart = e.Position
				c.Content.Refresh()
				return
			}
		}
	}
	if c.selectedCard != nil {
		c.selectedCard.SetSelected(false)
		c.selectedCard = nil
		c.Content.Refresh()
	}
}

func (c *MosugoCanvas) TappedSecondary(e *fyne.PointEvent) {
	c.isPanning = true
	c.panStart = e.Position
}

func (c *MosugoCanvas) Dragged(e *fyne.DragEvent) {
	if c.isPanning {
		delta := e.Position.Subtract(c.panStart)
		c.Offset.X += delta.X
		c.Offset.Y += delta.Y
		c.panStart = e.Position
		c.Refresh()
		return
	}

	if c.isDraggingCard && c.selectedCard != nil {
		delta := e.Position.Subtract(c.cardDragStart)
		c.selectedCard.WorldPos.X += delta.X
		c.selectedCard.WorldPos.Y += delta.Y
		c.cardDragStart = e.Position
		c.Refresh()
		return
	}

	switch c.CurrentTool {
	case tools.ToolDraw:
		if !c.isDrawing {
			c.isDrawing = true
			c.currentStroke = []*canvas.Line{}
			c.dragStart = e.Position
			c.lastDrawPos = e.Position
		}

		// Simple smoothing: Avoid drawing if movement is too small
		if math.Abs(float64(e.Position.X-c.lastDrawPos.X)) < 2 && math.Abs(float64(e.Position.Y-c.lastDrawPos.Y)) < 2 {
			return
		}

		prevPos := c.lastDrawPos
		c.lastDrawPos = e.Position

		// Linear Interpolation for smoothness
		dx := e.Position.X - prevPos.X
		dy := e.Position.Y - prevPos.Y
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

		// If moving fast, interpolate with fewer segments but enough to cover gaps
		// Using a dynamic step size based on speed might be better, but fixed step is fine for now
		if dist > 2 {
			line := canvas.NewLine(c.StrokeColor)
			line.StrokeWidth = c.StrokeWidth
			line.Position1 = prevPos
			line.Position2 = e.Position
			c.currentStroke = append(c.currentStroke, line)
			c.Content.Add(line)
		}

		c.Content.Refresh()
		return

	case tools.ToolErase:
		for _, obj := range c.Content.Objects {
			if mosuW, ok := obj.(*cards.MosuWidget); ok {
				screenPos := fyne.NewPos(
					float32(int(mosuW.WorldPos.X)+int(c.Offset.X)),
					float32(int(mosuW.WorldPos.Y)+int(c.Offset.Y)),
				)
				if e.Position.X >= screenPos.X && e.Position.X <= screenPos.X+mosuW.WorldSize.Width &&
					e.Position.Y >= screenPos.Y && e.Position.Y <= screenPos.Y+mosuW.WorldSize.Height {
					c.Content.Remove(mosuW)
					if c.selectedCard == mosuW {
						c.selectedCard = nil
					}
					c.Content.Refresh()
					return
				}
			} else if line, ok := obj.(*canvas.Line); ok {
				if c.pointNearLine(e.Position, line, 10) {
					c.Content.Remove(line)
					for i, stroke := range c.strokes {
						for j, l := range stroke {
							if l == line {
								c.strokes[i] = append(stroke[:j], stroke[j+1:]...)
								break
							}
						}
					}
					c.Content.Refresh()
					return
				}
			}
		}
		return

	case tools.ToolCard:
		if !c.isDragging {
			c.isDragging = true
			c.dragStart = e.Position.Subtract(e.Dragged)
			c.ghostRect.Show()
		}

		currPos := e.Position

		x1 := math.Min(float64(c.dragStart.X), float64(currPos.X))
		y1 := math.Min(float64(c.dragStart.Y), float64(currPos.Y))
		x2 := math.Max(float64(c.dragStart.X), float64(currPos.X))
		y2 := math.Max(float64(c.dragStart.Y), float64(currPos.Y))

		rawWorldX1 := float32(x1) - c.Offset.X
		rawWorldY1 := float32(y1) - c.Offset.Y
		rawWorldX2 := float32(x2) - c.Offset.X
		rawWorldY2 := float32(y2) - c.Offset.Y

		worldX1 := int(rawWorldX1)
		worldY1 := int(rawWorldY1)
		worldX2 := int(rawWorldX2)
		worldY2 := int(rawWorldY2)

		snapX1 := (worldX1 / GridSize) * GridSize
		snapY1 := (worldY1 / GridSize) * GridSize

		snapX2 := ((worldX2 / GridSize) + 1) * GridSize
		snapY2 := ((worldY2 / GridSize) + 1) * GridSize

		if snapX2 <= snapX1 {
			snapX2 = snapX1 + GridSize
		}
		if snapY2 <= snapY1 {
			snapY2 = snapY1 + GridSize
		}

		snapW := float32(snapX2 - snapX1)
		snapH := float32(snapY2 - snapY1)

		screenX := float32(snapX1 + int(c.Offset.X))
		screenY := float32(snapY1 + int(c.Offset.Y))

		c.ghostRect.Move(fyne.NewPos(screenX, screenY))
		c.ghostRect.Resize(fyne.NewSize(snapW, snapH))
		c.Content.Refresh()
	}
}

func (c *MosugoCanvas) DragEnd() {
	if c.isPanning {
		c.isPanning = false
		return
	}

	if c.isDraggingCard {
		c.isDraggingCard = false
		return
	}

	if c.isDrawing {
		c.isDrawing = false
		if len(c.currentStroke) > 0 {
			c.strokes = append(c.strokes, c.currentStroke)
			// Don't clear currentStroke here if it's reused, but we re-init it in Dragged start
			// Just nil it to be safe
			c.currentStroke = nil
		}
		return
	}

	if c.CurrentTool == tools.ToolCard && c.isDragging {
		rectSize := c.ghostRect.Size()
		rectPos := c.ghostRect.Position()

		c.ghostRect.Hide()
		c.isDragging = false

		worldW := int(rectSize.Width)
		worldH := int(rectSize.Height)
		worldX := int(rectPos.X - c.Offset.X)
		worldY := int(rectPos.Y - c.Offset.Y)

		if worldW < GridSize || worldH < GridSize {
			c.Content.Refresh()
			return
		}

		cardID := fmt.Sprintf("card_%d", len(c.Content.Objects))
		newCard := cards.NewMosuWidget(cardID, theme.CardWhite)

		newCard.WorldPos = fyne.NewPos(float32(worldX), float32(worldY))
		newCard.WorldSize = fyne.NewSize(float32(worldW), float32(worldH))

		screenX := float32(worldX + int(c.Offset.X))
		screenY := float32(worldY + int(c.Offset.Y))

		newCard.Move(fyne.NewPos(screenX, screenY))
		newCard.Resize(fyne.NewSize(float32(worldW), float32(worldH)))

		go c.animateCardFadeIn(newCard)

		c.Content.Add(newCard)
		c.Content.Refresh()
	}
}

func (c *MosugoCanvas) pointNearLine(pt fyne.Position, line *canvas.Line, threshold float32) bool {
	x0, y0 := pt.X, pt.Y
	x1, y1 := line.Position1.X, line.Position1.Y
	x2, y2 := line.Position2.X, line.Position2.Y

	dx := x2 - x1
	dy := y2 - y1
	lengthSq := dx*dx + dy*dy

	if lengthSq < 0.0001 {
		dist := float32(math.Sqrt(float64((x0-x1)*(x0-x1) + (y0-y1)*(y0-y1))))
		return dist <= threshold
	}

	t := ((x0-x1)*dx + (y0-y1)*dy) / lengthSq
	if t < 0 {
		t = 0
	} else if t > 1 {
		t = 1
	}

	nearX := x1 + t*dx
	nearY := y1 + t*dy
	dist := float32(math.Sqrt(float64((x0-nearX)*(x0-nearX) + (y0-nearY)*(y0-nearY))))

	return dist <= threshold
}

func (c *MosugoCanvas) animateCardFadeIn(card *cards.MosuWidget) {
	steps := 20
	duration := 200 * time.Millisecond
	stepDuration := duration / time.Duration(steps)

	originalSize := card.WorldSize
	for i := 0; i <= steps; i++ {
		progress := float32(i) / float32(steps)
		scale := 0.8 + 0.2*progress

		scaledW := originalSize.Width * scale
		scaledH := originalSize.Height * scale

		offsetX := (originalSize.Width - scaledW) / 2
		offsetY := (originalSize.Height - scaledH) / 2

		screenX := float32(int(card.WorldPos.X)+int(c.Offset.X)) + offsetX
		screenY := float32(int(card.WorldPos.Y)+int(c.Offset.Y)) + offsetY

		card.Move(fyne.NewPos(screenX, screenY))
		card.Resize(fyne.NewSize(scaledW, scaledH))
		card.Refresh()

		time.Sleep(stepDuration)
	}

	card.Resize(originalSize)
	screenX := float32(int(card.WorldPos.X) + int(c.Offset.X))
	screenY := float32(int(card.WorldPos.Y) + int(c.Offset.Y))
	card.Move(fyne.NewPos(screenX, screenY))
	card.Refresh()
}
