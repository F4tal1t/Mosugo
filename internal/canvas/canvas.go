package canvas

import (
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
	GridSize = 40
)

func snap(v float32) float32 {
	// Proper floor snapping for consistent grid alignment
	return float32(math.Floor(float64(v)/GridSize) * GridSize)
}

func snapUp(v float32) float32 {
	val := float64(v)
	snapped := math.Ceil(val/GridSize) * GridSize
	return float32(snapped)
}

func (c *MosugoCanvas) ScreenToWorld(pos fyne.Position) fyne.Position {
	x := (pos.X - c.Offset.X) / c.Scale
	y := (pos.Y - c.Offset.Y) / c.Scale
	return fyne.NewPos(x, y)
}

func (c *MosugoCanvas) WorldToScreen(pos fyne.Position) fyne.Position {
	x := (pos.X * c.Scale) + c.Offset.X
	y := (pos.Y * c.Scale) + c.Offset.Y
	return fyne.NewPos(x, y)
}

// --- tools.Canvas Implementation ---

func (c *MosugoCanvas) GetOffset() fyne.Position         { return c.Offset }
func (c *MosugoCanvas) SetOffset(pos fyne.Position)      { c.Offset = pos }
func (c *MosugoCanvas) GetScale() float32                { return c.Scale }
func (c *MosugoCanvas) AddObject(o fyne.CanvasObject)    { c.Content.Add(o) }
func (c *MosugoCanvas) RemoveObject(o fyne.CanvasObject) { c.Content.Remove(o) }
func (c *MosugoCanvas) SetCursor(cur desktop.Cursor) {
	// Fyne widgets don't have a specific "SetCursor" method in the API (cursor is queried)
	// We rely on the c.Cursor() method being called by the driver
	// Just requesting a refresh might be enough if the driver polls Cursor()
	c.Refresh()
}
func (c *MosugoCanvas) ContentObject() fyne.CanvasObject { return c.Content }

// Additional Canvas Interface Implementations

func (c *MosugoCanvas) Snap(v float32) float32                 { return snap(v) }
func (c *MosugoCanvas) SnapUp(v float32) float32               { return snapUp(v) }
func (c *MosugoCanvas) ContentContainer() *fyne.Container      { return c.Content }
func (c *MosugoCanvas) GhostRect() *canvas.Rectangle           { return c.ghostRect }
func (c *MosugoCanvas) GetSelectedCard() *cards.MosuWidget     { return c.selectedCard }
func (c *MosugoCanvas) SetSelectedCard(card *cards.MosuWidget) { c.selectedCard = card }

// SetTool is a helper to switch tools by enum (compatibility wrapper)
func (c *MosugoCanvas) SetTool(t tools.ToolType) {
	c.CurrentTool = t
	switch t {
	case tools.ToolPan:
		c.ActiveTool = &tools.PanTool{}
	case tools.ToolCard:
		c.ActiveTool = &tools.CardTool{}
	case tools.ToolDraw:
		c.ActiveTool = &tools.DrawTool{}
	case tools.ToolErase:
		c.ActiveTool = &tools.EraseTool{}
	// Add other tools here as implemented
	default:
		c.ActiveTool = &tools.CardTool{}
	}
	c.Refresh()
}

// -----------------------------------

type MosugoCanvas struct {
	widget.BaseWidget

	Grid    *canvas.Raster
	Content *fyne.Container

	Offset      fyne.Position
	Scale       float32
	DeviceScale float32
	// CurrentTool is kept for main.go compatibility, but logic delegates to ActiveTool
	CurrentTool tools.ToolType
	ActiveTool  tools.Tool

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

	// Map to store World Coordinates of strokes for rendering
	strokesMap map[*canvas.Line]StrokeCoords

	// Drawing smoothing
	lastDrawPos fyne.Position

	StrokeWidth float32
	StrokeColor color.Color
}

func NewMosugoCanvas() *MosugoCanvas {
	c := &MosugoCanvas{
		Offset:      fyne.NewPos(0, 0),
		Scale:       1.0,
		CurrentTool: tools.ToolCard,
		ActiveTool:  &tools.CardTool{}, // Default tool
		StrokeWidth: 2,
		StrokeColor: theme.InkGrey,
		strokesMap:  make(map[*canvas.Line]StrokeCoords),
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
	// Update device scale cache
	if drv := fyne.CurrentApp().Driver(); drv != nil {
		if cv := drv.CanvasForObject(r.canvas); cv != nil {
			r.canvas.DeviceScale = cv.Scale()
		} else {
			// If not attached yet, or no canvas, default or keep old.
			// But careful with 0 on first run.
			if r.canvas.DeviceScale == 0 {
				r.canvas.DeviceScale = 1.0
			}
		}
	}

	if r.canvas.Grid != nil {
		r.canvas.Grid.Resize(size)
		r.canvas.Grid.Move(fyne.NewPos(0, 0))
	}

	if r.canvas.Content != nil {
		r.canvas.Content.Resize(size)
		r.canvas.Content.Move(fyne.NewPos(0, 0))

		for _, obj := range r.canvas.Content.Objects {
			if mosuW, ok := obj.(*cards.MosuWidget); ok {
				// Use WorldToScreen for robust positioning (Float based)
				screenPos := r.canvas.WorldToScreen(mosuW.WorldPos)

				// Scale the size as well
				msgScale := r.canvas.Scale
				screenSize := fyne.NewSize(
					mosuW.WorldSize.Width*msgScale,
					mosuW.WorldSize.Height*msgScale,
				)

				mosuW.Move(screenPos)
				mosuW.Resize(screenSize)
			} else if rect, ok := obj.(*canvas.Rectangle); ok && rect == r.canvas.ghostRect {
				// Ghost rect logic handled by Tool
			} else if line, ok := obj.(*canvas.Line); ok {
				// Handle Lines (World -> Screen)
				// We need a map to store world coordinates. See MosugoCanvas struct.
				if coords, ok := r.canvas.GetStrokeCoords(line); ok {
					p1 := r.canvas.WorldToScreen(coords.P1)
					p2 := r.canvas.WorldToScreen(coords.P2)
					line.Position1 = p1
					line.Position2 = p2
					line.StrokeWidth = r.canvas.StrokeWidth * r.canvas.Scale
				}
			}
		}
	}
}

// ------------------------------------------------------------------
// Stroke Management Helpers
// ------------------------------------------------------------------

type StrokeCoords struct {
	P1, P2 fyne.Position
}

func (c *MosugoCanvas) AddStroke(p1, p2 fyne.Position) {
	line := canvas.NewLine(c.StrokeColor)
	line.StrokeWidth = c.StrokeWidth

	// Initial placement
	line.Position1 = c.WorldToScreen(p1)
	line.Position2 = c.WorldToScreen(p2)

	c.Content.Add(line)
	// Register logical coordinates
	c.RegisterStroke(line, p1, p2)
}

func (c *MosugoCanvas) RegisterStroke(line *canvas.Line, p1, p2 fyne.Position) {
	if c.strokesMap == nil {
		c.strokesMap = make(map[*canvas.Line]StrokeCoords)
	}
	c.strokesMap[line] = StrokeCoords{P1: p1, P2: p2}
}

func (c *MosugoCanvas) GetStrokeCoords(line *canvas.Line) (StrokeCoords, bool) {
	if c.strokesMap == nil {
		return StrokeCoords{}, false
	}
	coords, ok := c.strokesMap[line]
	return coords, ok
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
	if c.ActiveTool != nil {
		c.ActiveTool.OnTapped(c, e)
	}
}

func (c *MosugoCanvas) TappedSecondary(e *fyne.PointEvent) {
	// Usually secondary tap (right click) might just be context menu
	// But in some apps it pans.
	// For now, let's allow PanTool activation or just ignore if tool doesn't handle it
	// But our interface doesn't have OnTappedSecondary.
	// We can implement Pan interrupt here?
	c.isPanning = true // Keep old behavior or delegate?
	c.panStart = e.Position
	// Actually, let's invoke dragging logic if it moves?
	// But TappedSecondary is a click, not drag.
}

func (c *MosugoCanvas) Dragged(e *fyne.DragEvent) {
	// If right-mouse drag (often handled as Panning globally regardless of tool)
	// Fyne doesn't distinguish button in Dragged event easily unless we track MouseDown.
	// But c.isPanning is set by TappedSecondary?
	if c.isPanning {
		delta := e.Position.Subtract(c.panStart)
		c.Offset.X += delta.X
		c.Offset.Y += delta.Y
		c.panStart = e.Position
		c.Refresh()
		return
	}

	if c.ActiveTool != nil {
		c.ActiveTool.OnDragged(c, e)
	}
}

func (c *MosugoCanvas) DragEnd() {
	if c.isPanning {
		c.isPanning = false
		return
	}

	if c.ActiveTool != nil {
		c.ActiveTool.OnDragEnd(c)
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
