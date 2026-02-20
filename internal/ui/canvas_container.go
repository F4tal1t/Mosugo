package ui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"mosugo/internal/tools"
)

type invisibleTappable struct {
	widget.BaseWidget
	onTapped func()
}

func (t *invisibleTappable) Tapped(*fyne.PointEvent) {
	if t.onTapped != nil {
		t.onTapped()
	}
}

func (t *invisibleTappable) TappedSecondary(*fyne.PointEvent) {}

func (t *invisibleTappable) CreateRenderer() fyne.WidgetRenderer {
	rect := canvas.NewRectangle(color.Transparent)
	return widget.NewSimpleRenderer(rect)
}

func newInvisibleTappable(onTapped func()) *invisibleTappable {
	t := &invisibleTappable{onTapped: onTapped}
	t.ExtendBaseWidget(t)
	return t
}

// MetaballBorder is a custom widget wrapper that draws a metaball-like border
type MetaballBorder struct {
	widget.BaseWidget
	BorderColor color.Color
	Thickness   float32
	TabY        float32
	TabWidth    float32
	TabHeight   float32

	// Bottom tab for calendar
	BottomTabWidth     float32
	BottomTabHeight    float32
	BottomTabExpanded  bool
	BottomTabMaxHeight float32

	// Animation state
	currentBottomHeight float32
	animating           bool

	// Calendar content
	calendarContent fyne.CanvasObject
	dateLabel       *canvas.Text
	currentDate     time.Time

	// Tappable areas for calendar interaction
	calendarTabTapper *invisibleTappable
	backdropTapper    *invisibleTappable

	// Internal renderer reference for optimization
	renderer *metaballRenderer
}

func NewMetaballBorder(c color.Color) *MetaballBorder {
	m := &MetaballBorder{
		BorderColor:         c,
		Thickness:           10,
		TabY:                100,
		TabWidth:            60,
		TabHeight:           180,
		BottomTabWidth:      200,
		BottomTabHeight:     30,
		BottomTabMaxHeight:  240,
		currentBottomHeight: 30,
		BottomTabExpanded:   false,
		currentDate:         time.Now(),
	}

	// Create date label with theme color
	themeColor := color.RGBA{255, 255, 255, 255} // theme.InkWhite
	m.dateLabel = canvas.NewText(m.formatDate(m.currentDate), themeColor)
	m.dateLabel.Alignment = fyne.TextAlignCenter
	m.dateLabel.TextSize = 12

	// Create invisible tappable button for calendar tab (no hover color)
	m.calendarTabTapper = newInvisibleTappable(func() {
		m.ToggleCalendar()
	})

	// Create invisible backdrop tapper for closing calendar
	m.backdropTapper = newInvisibleTappable(func() {
		if m.BottomTabExpanded {
			m.ToggleCalendar()
		}
	})
	m.backdropTapper.Hide()

	m.ExtendBaseWidget(m)
	return m
}

// formatDate formats the date as "2026 / 02 / 19 - Thursday"
func (m *MetaballBorder) formatDate(date time.Time) string {
	return fmt.Sprintf("%d / %02d / %02d - %s",
		date.Year(),
		date.Month(),
		date.Day(),
		date.Weekday().String(),
	)
}

// SetCurrentDate updates the displayed date
func (m *MetaballBorder) SetCurrentDate(date time.Time) {
	m.currentDate = date
	m.dateLabel.Text = m.formatDate(date)
	m.Refresh()
}

// SetCalendarButtonSize allows resizing the calendar button dimensions
func (m *MetaballBorder) SetCalendarButtonSize(width, height, maxHeight float32) {
	m.BottomTabWidth = width
	m.BottomTabHeight = height
	m.BottomTabMaxHeight = maxHeight
	if !m.BottomTabExpanded {
		m.currentBottomHeight = height
	}
	m.Refresh()
}

// RefreshLayout updates positions during animation
func (m *MetaballBorder) RefreshLayout() {
	if m.renderer != nil {
		// Refresh raster to animate the border
		m.renderer.raster.Refresh()
		// Update layout positions
		m.renderer.Layout(m.Size())
		canvas.Refresh(m.renderer.calendarContainer)
		if m.dateLabel != nil {
			canvas.Refresh(m.dateLabel)
		}
		if m.calendarTabTapper != nil {
			canvas.Refresh(m.calendarTabTapper)
		}
		if m.backdropTapper != nil {
			canvas.Refresh(m.backdropTapper)
		}
	}
}

// SetCalendarContent sets the content to display in the expanded calendar tab
func (m *MetaballBorder) SetCalendarContent(content fyne.CanvasObject) {
	m.calendarContent = content
}

// ToggleCalendar expands or collapses the calendar tab with animation
func (m *MetaballBorder) ToggleCalendar() {
	if m.animating {
		return
	}

	m.animating = true
	m.BottomTabExpanded = !m.BottomTabExpanded

	targetHeight := m.BottomTabHeight
	if m.BottomTabExpanded {
		targetHeight = m.BottomTabMaxHeight
	}

	// Bounce animation using tools.BounceEasing
	startHeight := m.currentBottomHeight
	duration := 400 * time.Millisecond
	steps := 20 // Smooth animation with 20 steps

	go func() {
		// Enable low-resolution mode during animation
		if m.renderer != nil {
			m.renderer.isAnimating = true
		}

		for i := 0; i <= steps; i++ {
			progress := float32(i) / float32(steps)

			// Use BounceEasing from tools package for expanding only
			var t float32
			if m.BottomTabExpanded {
				t = tools.BounceEasing(progress)
			} else {
				// Smooth ease-out for collapsing (no bounce)
				t = progress
			}

			m.currentBottomHeight = startHeight + (targetHeight-startHeight)*t
			// Only refresh layout during animation, not the raster
			m.RefreshLayout()

			time.Sleep(duration / time.Duration(steps))
		}

		m.currentBottomHeight = targetHeight
		m.animating = false
		// Disable low-resolution mode for final high-quality render
		if m.renderer != nil {
			m.renderer.isAnimating = false
		}
		// Final refresh with raster update
		m.Refresh()
	}()
}

// IsPointInCalendarTab checks if a point is within the bottom calendar tab
func (m *MetaballBorder) IsPointInCalendarTab(pos fyne.Position, size fyne.Size) bool {
	// Calculate bottom tab bounds
	centerX := size.Width / 2
	bottomY := size.Height - m.Thickness - m.currentBottomHeight

	tabLeft := centerX - m.BottomTabWidth/2
	tabRight := centerX + m.BottomTabWidth/2
	tabTop := bottomY
	tabBottom := size.Height - m.Thickness

	return pos.X >= tabLeft && pos.X <= tabRight && pos.Y >= tabTop && pos.Y <= tabBottom
}

func (m *MetaballBorder) CreateRenderer() fyne.WidgetRenderer {
	r := &metaballRenderer{
		m: m,
	}
	r.raster = canvas.NewRaster(r.generator)
	r.calendarContainer = container.NewStack()
	m.renderer = r // Store reference for optimization
	return r
}

type metaballRenderer struct {
	m                 *MetaballBorder
	raster            *canvas.Raster
	calendarContainer *fyne.Container

	// Performance optimization: reusable buffer
	cachedImageBuffer *image.RGBA
	isAnimating       bool // Flag to use lower resolution during animation
}

func (r *metaballRenderer) MinSize() fyne.Size {
	return fyne.NewSize(100, 100)
}

func (r *metaballRenderer) Layout(s fyne.Size) {
	r.raster.Resize(s)

	centerX := s.Width / 2
	bottomY := s.Height - r.m.Thickness - r.m.currentBottomHeight

	// Position and show/hide backdrop tapper (full screen when calendar expanded)
	if r.m.BottomTabExpanded {
		r.m.backdropTapper.Resize(s)
		r.m.backdropTapper.Move(fyne.NewPos(0, 0))
		r.m.backdropTapper.Show()
		r.m.calendarTabTapper.Hide() // Hide when expanded
	} else {
		r.m.backdropTapper.Hide()
		// Position and show calendar tab tapper when collapsed
		r.m.calendarTabTapper.Resize(fyne.NewSize(r.m.BottomTabWidth, r.m.currentBottomHeight))
		r.m.calendarTabTapper.Move(fyne.NewPos(centerX-r.m.BottomTabWidth/2, bottomY))
		r.m.calendarTabTapper.Show()
	}

	// Position calendar content in the expanded bottom tab
	if r.m.calendarContent != nil && r.m.BottomTabExpanded {
		contentWidth := r.m.BottomTabWidth - 20
		contentHeight := r.m.currentBottomHeight - 20

		r.calendarContainer.Resize(fyne.NewSize(contentWidth, contentHeight))
		r.calendarContainer.Move(fyne.NewPos(
			centerX-contentWidth/2,
			bottomY,
		))

		r.calendarContainer.Objects = []fyne.CanvasObject{r.m.calendarContent}
		r.calendarContainer.Show()
	} else {
		r.calendarContainer.Hide()
	}

	if !r.m.BottomTabExpanded && r.m.dateLabel != nil {
		centerX := s.Width / 2
		bottomY := s.Height - r.m.Thickness - r.m.currentBottomHeight/2

		textSize := r.m.dateLabel.MinSize()
		r.m.dateLabel.Resize(textSize)
		r.m.dateLabel.Move(fyne.NewPos(
			centerX-textSize.Width/2,
			bottomY-textSize.Height/2,
		))
	}
}

func (r *metaballRenderer) Refresh() {
	r.raster.Refresh()
	r.Layout(r.m.Size())
	canvas.Refresh(r.calendarContainer)
}

func (r *metaballRenderer) Destroy() {}

func (r *metaballRenderer) Objects() []fyne.CanvasObject {
	if !r.m.BottomTabExpanded && r.m.dateLabel != nil {
		// Collapsed: show tab tapper and date label
		return []fyne.CanvasObject{r.raster, r.m.calendarTabTapper, r.m.dateLabel}
	}
	// Expanded: backdrop tapper behind, then calendar content on top
	return []fyne.CanvasObject{r.raster, r.m.backdropTapper, r.calendarContainer}
}

func sdBox(pX, pY, bX, bY float32) float32 {
	dX := float32(math.Abs(float64(pX))) - bX
	dY := float32(math.Abs(float64(pY))) - bY
	mX := float32(math.Max(float64(dX), 0))
	mY := float32(math.Max(float64(dY), 0))
	distOutside := float32(math.Sqrt(float64(mX*mX + mY*mY)))
	distInside := float32(math.Min(math.Max(float64(dX), float64(dY)), 0))
	return distOutside + distInside
}

func opSmoothUnion(d1, d2, k float32) float32 {
	h := float32(math.Max(float64(k)-math.Abs(float64(d1-d2)), 0.0)) / k
	return float32(math.Min(float64(d1), float64(d2))) - h*h*k*0.25
}

func (r *metaballRenderer) generator(w, h int) image.Image {
	// Use lower resolution during animation for performance
	renderW, renderH := w, h
	if r.isAnimating {
		// Render at 0.6x resolution during animation (2.7x faster)
		renderW = int(float32(w) * 0.6)
		renderH = int(float32(h) * 0.6)
		if renderW < 100 {
			renderW = 100
		}
		if renderH < 100 {
			renderH = 100
		}
	}

	// Reuse buffer if dimensions match, otherwise allocate new
	var img *image.RGBA
	if r.cachedImageBuffer != nil && r.cachedImageBuffer.Bounds().Dx() == renderW && r.cachedImageBuffer.Bounds().Dy() == renderH {
		img = r.cachedImageBuffer
		// Clear the buffer
		for i := range img.Pix {
			img.Pix[i] = 0
		}
	} else {
		img = image.NewRGBA(image.Rect(0, 0, renderW, renderH))
		r.cachedImageBuffer = img
	}

	col := r.m.BorderColor
	rInt, gInt, bInt, aInt := col.RGBA()

	scale := float32(1.0)
	sz := r.m.Size()
	if sz.Width > 0 && renderW > 0 {
		scale = float32(renderW) / sz.Width
	}

	thick := r.m.Thickness * scale
	tW := r.m.TabWidth * scale
	tH := r.m.TabHeight * scale
	fW, fH := float32(renderW), float32(renderH)

	// Bottom tab dimensions (scaled)
	bTabW := r.m.BottomTabWidth * scale
	bTabH := r.m.currentBottomHeight * scale

	// Create hole coordinates (inner rectangle)
	holeW := fW - 2*thick
	holeH := fH - 2*thick
	holeHalfW := holeW / 2
	holeHalfH := holeH / 2
	centerX := fW / 2
	centerY := fH / 2

	// Normalized Color components
	fR := float32(rInt) / 65535.0
	fG := float32(gInt) / 65535.0
	fB := float32(bInt) / 65535.0
	fA := float32(aInt) / 65535.0

	k := 20 * scale // smoothness factor

	// Left tab position (centered on the left inner border edge)
	leftTabCX := -holeHalfW
	leftTabCY := float32(0)

	// Bottom tab position (centered on bottom inner border edge)
	bottomTabCX := float32(0)
	bottomTabCY := holeHalfH

	for y := 0; y < renderH; y++ {
		for x := 0; x < renderW; x++ {
			pX := float32(x)
			pY := float32(y)

			// 1. SDF for Frame
			relX := pX - centerX
			relY := pY - centerY

			rad := 12 * scale
			dRectHole := sdBox(relX, relY, holeHalfW-rad, holeHalfH-rad) - rad
			sdFrame := -dRectHole // Negative when outside the hole (solid frame)

			// 2. SDF for Left Tab
			leftTabRad := 15 * scale
			sdLeftTab := sdBox(relX-leftTabCX, relY-leftTabCY, tW/1.4-leftTabRad, tH/2-leftTabRad) - leftTabRad

			// 3. SDF for Bottom Tab (calendar)
			bottomTabRad := 15 * scale
			sdBottomTab := sdBox(relX-bottomTabCX, relY-bottomTabCY, bTabW/2-bottomTabRad, bTabH-bottomTabRad) - bottomTabRad

			// 4. Union (merge frame, left tab, and bottom tab)
			dTemp := opSmoothUnion(sdFrame, sdLeftTab, k)
			dFinal := opSmoothUnion(dTemp, sdBottomTab, k)

			// 5. Render
			alpha := float32(0.0)
			val := dFinal

			if val < -0.5 {
				alpha = 1.0
			} else if val > 0.5 {
				alpha = 0.0
			} else {
				t := (val + 0.5)
				alpha = 1.0 - t
			}

			if alpha > 0 {
				offset := (y*renderW + x) * 4
				img.Pix[offset] = uint8(fR * alpha * 255)
				img.Pix[offset+1] = uint8(fG * alpha * 255)
				img.Pix[offset+2] = uint8(fB * alpha * 255)
				img.Pix[offset+3] = uint8(fA * alpha * 255)
			}
		}
	}
	return img
}
