package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"mosugo/internal/storage"
	"mosugo/internal/theme"
)

// tappableButton is a custom button with reduced height
type tappableButton struct {
	widget.BaseWidget
	onTapped  func()
	minHeight float32
}

func (t *tappableButton) Tapped(*fyne.PointEvent) {
	if t.onTapped != nil {
		t.onTapped()
	}
}

func (t *tappableButton) TappedSecondary(*fyne.PointEvent) {}

func (t *tappableButton) CreateRenderer() fyne.WidgetRenderer {
	rect := canvas.NewRectangle(color.Transparent)
	return widget.NewSimpleRenderer(rect)
}

func (t *tappableButton) MinSize() fyne.Size {
	return fyne.NewSize(30, t.minHeight)
}

// compactGridLayout is a custom grid layout with minimal vertical spacing
type compactGridLayout struct {
	cols            int
	verticalSpacing float32
}

func (g *compactGridLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	if len(objects) == 0 {
		return fyne.NewSize(0, 0)
	}

	rows := (len(objects) + g.cols - 1) / g.cols
	cellSize := objects[0].MinSize()

	width := float32(g.cols) * cellSize.Width
	height := float32(rows)*cellSize.Height + float32(rows-1)*g.verticalSpacing

	return fyne.NewSize(width, height)
}

func (g *compactGridLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	if len(objects) == 0 {
		return
	}

	cellWidth := size.Width / float32(g.cols)
	// Use a consistent cell height instead of relying on MinSize
	cellHeight := float32(22) // Fixed height for calendar day cells

	for i, obj := range objects {
		row := i / g.cols
		col := i % g.cols

		x := float32(col) * cellWidth
		y := float32(row) * (cellHeight + g.verticalSpacing)

		obj.Resize(fyne.NewSize(cellWidth, cellHeight))
		obj.Move(fyne.NewPos(x, y))
	}
}

func newCompactGrid(cols int, verticalSpacing float32, objects ...fyne.CanvasObject) *fyne.Container {
	return container.New(&compactGridLayout{cols: cols, verticalSpacing: verticalSpacing}, objects...)
}

// CalendarContent creates the calendar month view content for embedding
type CalendarContent struct {
	widget.BaseWidget

	currentDate    time.Time
	onDateSelected func(time.Time)

	dateLabel *widget.Label
	monthView *fyne.Container
	renderer  *calendarRenderer
}

func NewCalendarContent(currentDate time.Time, onDateSelected func(time.Time)) *CalendarContent {
	c := &CalendarContent{
		currentDate:    currentDate,
		onDateSelected: onDateSelected,
	}
	c.ExtendBaseWidget(c)
	c.buildContent()
	return c
}

func (c *CalendarContent) buildContent() {
	// Date display at top (convert to canvas.Text)
	c.dateLabel = widget.NewLabel(c.formatDate(c.currentDate))
	c.dateLabel.Alignment = fyne.TextAlignCenter
	c.dateLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Build month view
	c.monthView = c.createMonthView(c.currentDate)
}

func (c *CalendarContent) formatDate(date time.Time) string {
	return fmt.Sprintf("%d / %02d / %02d - %s",
		date.Year(),
		date.Month(),
		date.Day(),
		date.Weekday().String(),
	)
}

func (c *CalendarContent) createMonthView(date time.Time) *fyne.Container {
	// Get saved dates
	savedDates, err := storage.ListSavedDates()
	if err != nil {
		savedDates = []time.Time{}
	}

	savedDatesMap := make(map[string]bool)
	for _, d := range savedDates {
		savedDatesMap[d.Format("2006-01-02")] = true
	}

	firstDay := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	lastDay := firstDay.AddDate(0, 1, -1)

	// Header with weekdays (single letters)
	weekdayLabels := []string{"S", "M", "T", "W", "T", "F", "S"}
	var headerObjects []fyne.CanvasObject
	for _, day := range weekdayLabels {
		label := canvas.NewText(day, theme.InkLightGrey)
		label.Alignment = fyne.TextAlignCenter
		label.TextStyle = fyne.TextStyle{Bold: true}
		headerObjects = append(headerObjects, label)
	}
	headerRow := newCompactGrid(7, 2, headerObjects...)

	// Days grid with reduced vertical spacing
	var dayObjects []fyne.CanvasObject

	// Empty cells before month starts
	firstWeekday := int(firstDay.Weekday())
	for i := 0; i < firstWeekday; i++ {
		dayObjects = append(dayObjects, layout.NewSpacer())
	}

	// Day buttons
	for day := 1; day <= lastDay.Day(); day++ {
		dayDate := time.Date(date.Year(), date.Month(), day, 0, 0, 0, 0, date.Location())
		dateStr := dayDate.Format("2006-01-02")
		hasSavedWork := savedDatesMap[dateStr]

		dayButton := c.createDayButton(day, dayDate, hasSavedWork)
		dayObjects = append(dayObjects, dayButton)
	}

	daysGrid := newCompactGrid(7, 2, dayObjects...)

	// Month/Year label above navigation (format: "Jan 26", "Feb 26", "Mar 26")
	monthYearText := canvas.NewText(date.Format("Jan 06"), theme.InkWhite)
	monthYearText.Alignment = fyne.TextAlignCenter
	monthYearText.TextStyle = fyne.TextStyle{Bold: true}
	monthYearText.TextSize = 14

	// Navigation buttons using actual Fyne buttons with no background
	prevButton := widget.NewButton("<", func() {
		c.currentDate = c.currentDate.AddDate(0, -1, 0)
		c.buildContent()
		c.BaseWidget.Refresh()
	})
	prevButton.Importance = widget.LowImportance

	nextButton := widget.NewButton(">", func() {
		c.currentDate = c.currentDate.AddDate(0, 1, 0)
		c.buildContent()
		c.BaseWidget.Refresh()
	})
	nextButton.Importance = widget.LowImportance

	todayButton := widget.NewButton("Today", func() {
		if c.onDateSelected != nil {
			c.onDateSelected(time.Now())
		}
	})
	todayButton.Importance = widget.LowImportance

	navButtons := container.NewBorder(nil, nil, prevButton, nextButton, container.NewCenter(todayButton))
	navRow := container.NewVBox(monthYearText, navButtons)

	// Add minimal spacing
	topSpacer := canvas.NewRectangle(color.Transparent)
	topSpacer.SetMinSize(fyne.NewSize(0, 5))

	return container.NewBorder(
		container.NewVBox(topSpacer, navRow, headerRow),
		nil, nil, nil,
		daysGrid,
	)
}

func (c *CalendarContent) createDayButton(day int, date time.Time, hasSavedWork bool) fyne.CanvasObject {
	dayStr := fmt.Sprintf("%d", day)

	// Determine color based on date state
	today := time.Now()
	isToday := date.Year() == today.Year() && date.Month() == today.Month() && date.Day() == today.Day()

	var textColor color.Color
	if isToday {
		textColor = theme.CardTurquoise
	} else if hasSavedWork {
		textColor = color.RGBA{164, 159, 234, 255}
	} else {
		textColor = theme.InkLightGrey
	}

	// Create text with determined color
	text := canvas.NewText(dayStr, textColor)
	text.Alignment = fyne.TextAlignCenter
	text.TextSize = 16

	// Create tappable button with smaller height
	tappable := &tappableButton{
		onTapped: func() {
			if c.onDateSelected != nil {
				c.onDateSelected(date)
			}
		},
		minHeight: 18,
	}
	tappable.ExtendBaseWidget(tappable)

	// Overlay text on button
	return container.NewStack(tappable, container.NewCenter(text))
}

func (c *CalendarContent) Refresh() {
	c.buildContent()
	if c.renderer != nil {
		c.renderer.Refresh()
	}
	c.BaseWidget.Refresh()
}

func (c *CalendarContent) GetCurrentDate() time.Time {
	return c.currentDate
}

func (c *CalendarContent) SetCurrentDate(date time.Time) {
	c.currentDate = date
	c.buildContent()
	if c.renderer != nil {
		c.renderer.Refresh()
	}
	c.BaseWidget.Refresh()
}

type calendarRenderer struct {
	content *CalendarContent
}

func (r *calendarRenderer) Destroy() {}

func (r *calendarRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.content.monthView}
}

func (r *calendarRenderer) Refresh() {
	if r.content.monthView != nil {
		canvas.Refresh(r.content.monthView)
	}
}

func (r *calendarRenderer) Layout(size fyne.Size) {
	if r.content.monthView != nil {
		r.content.monthView.Resize(size)
	}
}

func (r *calendarRenderer) MinSize() fyne.Size {
	if r.content.monthView != nil {
		return r.content.monthView.MinSize()
	}
	return fyne.NewSize(0, 0)
}

func (c *CalendarContent) CreateRenderer() fyne.WidgetRenderer {
	c.renderer = &calendarRenderer{content: c}
	return c.renderer
}
