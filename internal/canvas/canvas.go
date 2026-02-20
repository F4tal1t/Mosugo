// Package canvas provides the infinite zoomable canvas implementation for Mosugo.
// It handles coordinate transformations between world space (infinite canvas coordinates)
// and screen space (viewport pixels), manages pan/zoom state, and provides the main
// drawing surface for cards and freehand strokes.
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
	"mosugo/internal/storage"
	"mosugo/internal/theme"
	"mosugo/internal/tools"
)

const (
	GridSize = 30
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

func (c *MosugoCanvas) GetOffset() fyne.Position      { return c.Offset }
func (c *MosugoCanvas) SetOffset(pos fyne.Position)   { c.Offset = pos }
func (c *MosugoCanvas) GetScale() float32             { return c.Scale }
func (c *MosugoCanvas) AddObject(o fyne.CanvasObject) { c.Content.Add(o) }
func (c *MosugoCanvas) RemoveObject(o fyne.CanvasObject) {
	if line, ok := o.(*canvas.Line); ok {
		delete(c.strokesMap, line)
		delete(c.strokeIDMap, line)
		delete(c.glowLines, line)
	}
	c.Content.Remove(o)
}

func (c *MosugoCanvas) SetCursor(cur desktop.Cursor) {

	c.Refresh()
}

func (c *MosugoCanvas) ContentObject() fyne.CanvasObject       { return c.Content }
func (c *MosugoCanvas) Snap(v float32) float32                 { return snap(v) }
func (c *MosugoCanvas) SnapUp(v float32) float32               { return snapUp(v) }
func (c *MosugoCanvas) ContentContainer() *fyne.Container      { return c.Content }
func (c *MosugoCanvas) GhostRect() *canvas.Rectangle           { return c.ghostRect }
func (c *MosugoCanvas) GetSelectedCard() *cards.MosuWidget     { return c.selectedCard }
func (c *MosugoCanvas) SetSelectedCard(card *cards.MosuWidget) { c.selectedCard = card }
func (c *MosugoCanvas) SetTool(t tools.ToolType) {
	c.CurrentTool = t
	switch t {
	case tools.ToolSelect:
		c.ActiveTool = &tools.SelectTool{}
	case tools.ToolCard:
		c.ActiveTool = &tools.CardTool{}
	case tools.ToolDraw:
		c.ActiveTool = &tools.DrawTool{}
	case tools.ToolErase:
		c.ActiveTool = &tools.EraseTool{}
	default:
		c.ActiveTool = &tools.SelectTool{}
	}
	c.Refresh()
}

// -----------------------------------

// MosugoCanvas is the main infinite zoomable canvas widget.
// It handles coordinate transformations between world space and screen space,
// manages cards and strokes, and delegates tool interactions to the active tool.
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

	ghostRect    *canvas.Rectangle
	selectedCard *cards.MosuWidget

	isPanning bool
	panStart  fyne.Position

	currentStroke []*canvas.Line
	strokes       [][]*canvas.Line

	strokesMap   map[*canvas.Line]StrokeCoords
	strokeIDMap  map[*canvas.Line]int
	glowLines    map[*canvas.Line]bool
	nextStrokeID int

	StrokeWidth float32
	StrokeColor color.Color

	lastScale float32

	// Persistence fields
	currentDate time.Time
	isDirty     bool
	onDirty     func() // Callback when canvas becomes dirty
}

// NewMosugoCanvas creates and initializes a new MosugoCanvas with default settings.
// The canvas starts at 1:1 scale with zero offset and the Select tool active.
func NewMosugoCanvas() *MosugoCanvas {
	c := &MosugoCanvas{
		Offset:       fyne.NewPos(0, 0),
		Scale:        1.0,
		CurrentTool:  tools.ToolSelect,
		ActiveTool:   &tools.SelectTool{},
		StrokeWidth:  2.5,
		StrokeColor:  theme.InkGrey,
		strokesMap:   make(map[*canvas.Line]StrokeCoords),
		strokeIDMap:  make(map[*canvas.Line]int),
		glowLines:    make(map[*canvas.Line]bool),
		nextStrokeID: 1,
		currentDate:  time.Now(),
		isDirty:      false,
	}
	c.ExtendBaseWidget(c)

	c.Grid = BoxGridPattern(c, GridSize, theme.GridLine, theme.GridBg)
	c.Content = container.NewWithoutLayout()

	c.ghostRect = canvas.NewRectangle(theme.GridBg)
	c.ghostRect.StrokeColor = theme.GridLine
	c.ghostRect.StrokeWidth = 2
	c.ghostRect.Hide()

	c.Content.Add(c.ghostRect)

	return c
}

// CreateRenderer creates and returns the widget renderer for the canvas.
func (c *MosugoCanvas) CreateRenderer() fyne.WidgetRenderer {
	return &mosugoRenderer{canvas: c}
}

type mosugoRenderer struct {
	canvas *MosugoCanvas
}

// Destroy cleans up the renderer resources.
func (r *mosugoRenderer) Destroy() {}

// Layout positions and sizes all canvas objects based on the current view.
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
		gSize := float64(GridSize) * float64(r.canvas.Scale)

		if gSize < 1 {
			gSize = 1
		}

		shiftX := math.Mod(float64(r.canvas.Offset.X), gSize)
		shiftY := math.Mod(float64(r.canvas.Offset.Y), gSize)

		posX := shiftX - gSize
		posY := shiftY - gSize

		newSize := size.Add(fyne.NewSize(float32(gSize*2), float32(gSize*2)))
		r.canvas.Grid.Resize(newSize)
		r.canvas.Grid.Move(fyne.NewPos(float32(posX), float32(posY)))
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
				if coords, ok := r.canvas.GetStrokeCoords(line); ok {
					p1 := r.canvas.WorldToScreen(coords.P1)
					p2 := r.canvas.WorldToScreen(coords.P2)
					line.Position1 = p1
					line.Position2 = p2

					if r.canvas.glowLines[line] {
						line.StrokeWidth = r.canvas.StrokeWidth * 2 * r.canvas.Scale
					} else {
						line.StrokeWidth = r.canvas.StrokeWidth * r.canvas.Scale
					}
				}
			}
		}
	}
}

// ------------------------------------------------------------------
// Stroke Management Helpers
// ------------------------------------------------------------------

// StrokeCoords holds the world-space start and end positions of a stroke line.
type StrokeCoords struct {
	P1, P2 fyne.Position
}

// AddStroke adds a new stroke line to the canvas with the specified coordinates and stroke ID.
func (c *MosugoCanvas) AddStroke(p1, p2 fyne.Position, strokeID int) {
	// Defensive check: ensure stroke ID is valid
	if !c.ValidateStrokeID(strokeID) {
		// This shouldn't happen, but generate a valid ID if it does
		strokeID = c.GenerateStrokeID()
	}

	glowLine := canvas.NewLine(theme.GridBg)
	glowLine.StrokeWidth = c.StrokeWidth * 1.5
	glowLine.Position1 = c.WorldToScreen(p1)
	glowLine.Position2 = c.WorldToScreen(p2)
	c.Content.Add(glowLine)
	c.RegisterStroke(glowLine, p1, p2, strokeID)
	c.glowLines[glowLine] = true

	line := canvas.NewLine(c.StrokeColor)
	line.StrokeWidth = c.StrokeWidth
	line.Position1 = c.WorldToScreen(p1)
	line.Position2 = c.WorldToScreen(p2)
	c.Content.Add(line)
	c.RegisterStroke(line, p1, p2, strokeID)
}

// RegisterStroke associates stroke metadata with a line object for tracking.
func (c *MosugoCanvas) RegisterStroke(line *canvas.Line, p1, p2 fyne.Position, strokeID int) {
	if c.strokesMap == nil {
		c.strokesMap = make(map[*canvas.Line]StrokeCoords)
	}
	if c.strokeIDMap == nil {
		c.strokeIDMap = make(map[*canvas.Line]int)
	}
	c.strokesMap[line] = StrokeCoords{P1: p1, P2: p2}
	c.strokeIDMap[line] = strokeID
}

// GetStrokeCoords retrieves the world coordinates for a given stroke line.
func (c *MosugoCanvas) GetStrokeCoords(line *canvas.Line) (StrokeCoords, bool) {
	if c.strokesMap == nil {
		return StrokeCoords{}, false
	}
	coords, ok := c.strokesMap[line]
	return coords, ok
}

// GetStrokePoints retrieves the start and end points of a stroke line in world coordinates.
func (c *MosugoCanvas) GetStrokePoints(line *canvas.Line) (fyne.Position, fyne.Position, bool) {
	if c.strokesMap == nil {
		return fyne.Position{}, fyne.Position{}, false
	}
	coords, ok := c.strokesMap[line]
	return coords.P1, coords.P2, ok
}

func (c *MosugoCanvas) GetStrokeID(line *canvas.Line) (int, bool) {
	if c.strokeIDMap == nil {
		return 0, false
	}
	id, ok := c.strokeIDMap[line]
	return id, ok
}

func (c *MosugoCanvas) IsGlowLine(line *canvas.Line) bool {
	if c.glowLines == nil {
		return false
	}
	return c.glowLines[line]
}

// GenerateStrokeID creates a unique ID for a new stroke.
func (c *MosugoCanvas) GenerateStrokeID() int {
	id := c.nextStrokeID
	c.nextStrokeID++
	return id
}

// ValidateStrokeID checks if a stroke ID is valid (non-zero)
func (c *MosugoCanvas) ValidateStrokeID(strokeID int) bool {
	return strokeID > 0
}

// SimplifyStroke uses Douglas-Peucker algorithm to reduce points while preserving shape
// epsilon controls simplification aggressiveness (3.0 is recommended for good balance)
func (c *MosugoCanvas) SimplifyStroke(points []fyne.Position, epsilon float32) []fyne.Position {
	if len(points) < 3 {
		return points
	}

	// Find point with maximum distance from line
	dmax := float32(0)
	index := 0
	end := len(points) - 1

	for i := 1; i < end; i++ {
		d := perpendicularDistance(points[i], points[0], points[end])
		if d > dmax {
			index = i
			dmax = d
		}
	}

	// If max distance > epsilon, recursively simplify
	if dmax > epsilon {
		left := c.SimplifyStroke(points[:index+1], epsilon)
		right := c.SimplifyStroke(points[index:], epsilon)
		return append(left[:len(left)-1], right...)
	}

	return []fyne.Position{points[0], points[end]}
}

// perpendicularDistance calculates perpendicular distance from point to line segment
func perpendicularDistance(point, lineStart, lineEnd fyne.Position) float32 {
	dx := lineEnd.X - lineStart.X
	dy := lineEnd.Y - lineStart.Y

	// Calculate magnitude
	mag := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if mag < 0.0001 {
		return distance(point, lineStart)
	}

	// Normalize
	dx /= mag
	dy /= mag

	// Vector from lineStart to point
	pvx := point.X - lineStart.X
	pvy := point.Y - lineStart.Y

	// Get dot product (project pv onto normalized line)
	dot := pvx*dx + pvy*dy

	// Clamp to segment
	if dot < 0 {
		return distance(point, lineStart)
	} else if dot > mag {
		return distance(point, lineEnd)
	}

	// Perpendicular distance
	return float32(math.Abs(float64(pvx*(-dy) + pvy*dx)))
}

// distance calculates Euclidean distance between two points
func distance(p1, p2 fyne.Position) float32 {
	dx := p2.X - p1.X
	dy := p2.Y - p1.Y
	return float32(math.Sqrt(float64(dx*dx + dy*dy)))
}

// MinSize returns the minimum size for the canvas renderer.
func (r *mosugoRenderer) MinSize() fyne.Size {
	return fyne.NewSize(100, 100)
}

// Objects returns the list of drawable objects for the canvas.
func (r *mosugoRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.canvas.Grid, r.canvas.Content}
}

// Refresh triggers a redraw of the canvas.
func (r *mosugoRenderer) Refresh() {
	if r.canvas.Grid != nil {
		if r.canvas.lastScale != r.canvas.Scale {
			r.canvas.lastScale = r.canvas.Scale
			r.canvas.Grid.Refresh()
		}
	}

	r.Layout(r.canvas.Size())
	canvas.Refresh(r.canvas.Content)
}

// Cursor returns the appropriate cursor style based on the active tool.
func (c *MosugoCanvas) Cursor() desktop.Cursor {
	switch c.CurrentTool {
	case tools.ToolCard:
		return desktop.CrosshairCursor
	case tools.ToolDraw:
		return desktop.TextCursor
	case tools.ToolErase:
		return desktop.HResizeCursor
	}
	return desktop.DefaultCursor
}

// Tapped handles single tap/click events on the canvas.
func (c *MosugoCanvas) Tapped(e *fyne.PointEvent) {
	if c.ActiveTool != nil {
		c.ActiveTool.OnTapped(c, e)
	}
}

// MouseDown handles mouse button down events (for middle-click pan).
func (c *MosugoCanvas) MouseDown(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonSecondary {
		c.isPanning = true
		c.panStart = e.Position
	}
}

// MouseUp handles mouse button release events.
func (c *MosugoCanvas) MouseUp(e *desktop.MouseEvent) {
	if e.Button == desktop.MouseButtonSecondary {
		c.isPanning = false
	}
}

// Dragged handles drag events and delegates to the active tool.
func (c *MosugoCanvas) Dragged(e *fyne.DragEvent) {
	// Panning takes precedence if right-mouse is held
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

// DragEnd handles the end of drag events.
func (c *MosugoCanvas) DragEnd() {
	if c.isPanning {
		c.isPanning = false
		return
	}

	if c.ActiveTool != nil {
		c.ActiveTool.OnDragEnd(c)
	}
}

// --- Persistence Methods ---

// MarkDirty marks the canvas as modified and triggers the dirty callback
func (c *MosugoCanvas) MarkDirty() {
	c.isDirty = true
	if c.onDirty != nil {
		c.onDirty()
	}
}

// SetOnDirty sets the callback function to be called when the canvas becomes dirty
func (c *MosugoCanvas) SetOnDirty(callback func()) {
	c.onDirty = callback
}

// GetCurrentDate returns the date of the currently loaded workspace
func (c *MosugoCanvas) GetCurrentDate() time.Time {
	return c.currentDate
}

// SetCurrentDate sets the date of the workspace being worked on
func (c *MosugoCanvas) SetCurrentDate(date time.Time) {
	c.currentDate = date
}

// SaveCurrentWorkspace saves the current canvas state to storage
func (c *MosugoCanvas) SaveCurrentWorkspace() error {
	state := storage.WorkspaceState{
		Scale:   c.Scale,
		OffsetX: c.Offset.X,
		OffsetY: c.Offset.Y,
		Cards:   []storage.MosuData{},
		Strokes: []storage.StrokeData{},
		Date:    c.currentDate.Format("2006-01-02"),
	}

	// Collect cards
	for _, obj := range c.Content.Objects {
		if card, ok := obj.(*cards.MosuWidget); ok {
			cardData := storage.MosuData{
				ID:        card.ID,
				Content:   card.GetText(),
				PosX:      card.WorldPos.X,
				PosY:      card.WorldPos.Y,
				Width:     card.WorldSize.Width,
				Height:    card.WorldSize.Height,
				ColorIdx:  card.ColorIndex,
				CreatedAt: card.CreatedAt,
			}
			state.Cards = append(state.Cards, cardData)
		}
	}

	// Collect strokes
	for line, coords := range c.strokesMap {
		// Skip glow lines (only save the actual strokes)
		if c.glowLines[line] {
			continue
		}

		// Validate stroke ID before saving
		strokeID, ok := c.GetStrokeID(line)
		if !ok || !c.ValidateStrokeID(strokeID) {
			// Skip strokes without valid IDs (shouldn't happen, but defensive)
			continue
		}

		strokeData := storage.StrokeData{
			P1X:      coords.P1.X,
			P1Y:      coords.P1.Y,
			P2X:      coords.P2.X,
			P2Y:      coords.P2.Y,
			ColorIdx: 0, // Default color index for now
			Width:    c.StrokeWidth,
			StrokeID: strokeID,
		}
		state.Strokes = append(state.Strokes, strokeData)
	}

	// Save to file
	if err := storage.SaveWorkspace(c.currentDate, state); err != nil {
		return err
	}

	c.isDirty = false
	return nil
}

// LoadWorkspace loads a workspace from storage and replaces the current canvas state
func (c *MosugoCanvas) LoadWorkspace(date time.Time) error {
	// Load workspace state
	state, err := storage.LoadWorkspace(date)
	if err != nil {
		return err
	}

	// Clear current canvas
	c.ClearCanvas()

	// Restore canvas state
	c.Scale = state.Scale
	c.Offset = fyne.NewPos(state.OffsetX, state.OffsetY)
	c.currentDate = date

	// Restore cards
	for _, cardData := range state.Cards {
		card := cards.NewMosuWidget(cardData.ID, theme.CardBg, cardData.ColorIdx)
		card.WorldPos = fyne.NewPos(cardData.PosX, cardData.PosY)
		card.WorldSize = fyne.NewSize(cardData.Width, cardData.Height)
		card.CreatedAt = cardData.CreatedAt
		card.SetText(cardData.Content)
		card.RefreshContent()
		c.Content.Add(card)
	}

	// Restore strokes
	maxStrokeID := 0
	for _, strokeData := range state.Strokes {
		p1 := fyne.NewPos(strokeData.P1X, strokeData.P1Y)
		p2 := fyne.NewPos(strokeData.P2X, strokeData.P2Y)

		// Migrate invalid stroke IDs (backward compatibility)
		strokeID := strokeData.StrokeID
		if !c.ValidateStrokeID(strokeID) {
			// Assign new valid ID for corrupted/old strokes
			strokeID = c.GenerateStrokeID()
		}

		c.AddStroke(p1, p2, strokeID)

		// Track maximum stroke ID for next ID generation
		if strokeID > maxStrokeID {
			maxStrokeID = strokeID
		}
	}

	// Set next stroke ID to avoid collisions
	if maxStrokeID >= c.nextStrokeID {
		c.nextStrokeID = maxStrokeID + 1
	}

	c.isDirty = false
	c.Refresh()
	return nil
}

// ClearCanvas removes all cards and strokes from the canvas.
func (c *MosugoCanvas) ClearCanvas() {
	objectsToRemove := []fyne.CanvasObject{}
	for _, obj := range c.Content.Objects {
		if obj != c.ghostRect {
			objectsToRemove = append(objectsToRemove, obj)
		}
	}

	for _, obj := range objectsToRemove {
		c.Content.Remove(obj)
	}

	// Clear stroke maps
	c.strokesMap = make(map[*canvas.Line]StrokeCoords)
	c.strokeIDMap = make(map[*canvas.Line]int)
	c.glowLines = make(map[*canvas.Line]bool)
	c.strokes = [][]*canvas.Line{}
	c.currentStroke = []*canvas.Line{}
	c.nextStrokeID = 1

	c.selectedCard = nil
}
