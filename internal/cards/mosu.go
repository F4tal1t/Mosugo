package cards

import (
	"image/color"

	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"

	"mosugo/internal/theme"
)

type MosuWidget struct {
	widget.BaseWidget
	ID string

	WorldPos  fyne.Position
	WorldSize fyne.Size

	bg         *canvas.Rectangle
	content    *widget.Entry
	isSelected bool
}

func NewMosuWidget(id string, c color.Color) *MosuWidget {
	m := &MosuWidget{ID: id}
	m.ExtendBaseWidget(m)

	m.bg = canvas.NewRectangle(c)
	m.bg.CornerRadius = 0
	m.bg.StrokeColor = color.RGBA{0, 0, 0, 0}
	m.bg.StrokeWidth = 0

	m.content = widget.NewMultiLineEntry()
	m.content.Wrapping = fyne.TextWrapWord
	m.content.SetPlaceHolder("Type your note...")

	m.content.OnChanged = func(s string) {
		start := m.content.CursorColumn
		row := m.content.CursorRow

		if strings.Contains(s, "[] ") {
			newText := strings.ReplaceAll(s, "[] ", "☐ ")
			if newText != s {
				m.content.SetText(newText)
				m.content.CursorColumn = start - 2
				if m.content.CursorColumn < 0 {
					m.content.CursorColumn = 0
				}
				m.content.CursorRow = row
			}
		}

		if strings.Contains(s, "[x] ") {
			newText := strings.ReplaceAll(s, "[x] ", "☑ ")
			if newText != s {
				m.content.SetText(newText)
			}
		}
	}

	m.content.Resize(fyne.NewSize(100, 100))

	return m
}

func (m *MosuWidget) FocusEntry() {
	if m.content != nil {
		m.content.FocusGained()
	}
}

func (m *MosuWidget) SetSelected(selected bool) {
	m.isSelected = selected
	if selected {
		m.bg.StrokeColor = theme.SelectionBlue
		m.bg.StrokeWidth = 3
	} else {
		m.bg.StrokeColor = color.RGBA{0, 0, 0, 0}
		m.bg.StrokeWidth = 0
	}
	m.Refresh()
}

func (m *MosuWidget) CreateRenderer() fyne.WidgetRenderer {
	return &mosuRenderer{
		mosu:    m,
		objects: []fyne.CanvasObject{m.bg, m.content},
	}
}

type mosuRenderer struct {
	mosu    *MosuWidget
	objects []fyne.CanvasObject
}

func (r *mosuRenderer) Layout(size fyne.Size) {
	r.mosu.bg.Move(fyne.NewPos(0, 0))
	r.mosu.bg.Resize(size)

	r.mosu.content.Move(fyne.NewPos(0, 0))
	r.mosu.content.Resize(size)
}

func (r *mosuRenderer) MinSize() fyne.Size {
	return fyne.NewSize(40, 40)
}

func (r *mosuRenderer) Refresh() {
	canvas.Refresh(r.mosu.bg)
	canvas.Refresh(r.mosu.content)
}

func (r *mosuRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *mosuRenderer) Destroy() {}

func (m *MosuWidget) RefreshLayout(zoom float32) {
	width := m.WorldSize.Width * zoom
	height := m.WorldSize.Height * zoom
	x := m.WorldPos.X * zoom
	y := m.WorldPos.Y * zoom

	m.Move(fyne.NewPos(x, y))
	m.Resize(fyne.NewSize(width, height))
}
