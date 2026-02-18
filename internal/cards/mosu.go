package cards

import (
	"image/color"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"mosugo/internal/theme"
)

// MosuWidget represents a card on the canvas
type MosuWidget struct {
	widget.BaseWidget
	ID string

	WorldPos  fyne.Position
	WorldSize fyne.Size

	bg          *canvas.Rectangle
	contentVBox *fyne.Container
	container   *fyne.Container

	isSelected bool
	rawText    string
}

func NewMosuWidget(id string, c color.Color) *MosuWidget {
	m := &MosuWidget{ID: id}
	m.ExtendBaseWidget(m)

	m.bg = canvas.NewRectangle(c)
	m.bg.CornerRadius = 3
	m.bg.StrokeColor = color.Transparent
	m.bg.StrokeWidth = 0
	m.bg.FillColor = c

	m.rawText = ""
	m.contentVBox = container.NewVBox()
	m.container = container.NewStack(m.bg, container.NewPadded(m.contentVBox))

	return m
}

// customCheck is a circular checkbox widget
type customCheck struct {
	widget.BaseWidget
	Label        string
	Checked      bool
	OnTappedFunc func(bool)
}

func newCustomCheck(label string, checked bool, cb func(bool)) *customCheck {
	c := &customCheck{
		Label:        label,
		Checked:      checked,
		OnTappedFunc: cb,
	}
	c.ExtendBaseWidget(c)
	return c
}

func (c *customCheck) CreateRenderer() fyne.WidgetRenderer {
	ring := canvas.NewCircle(color.Transparent)
	ring.StrokeColor = theme.InkGrey
	ring.StrokeWidth = 2

	dot := canvas.NewCircle(theme.InkGrey)
	if !c.Checked {
		dot.Hide()
	}

	label := widget.NewLabel(c.Label)
	label.Wrapping = fyne.TextWrapWord

	return &customCheckRenderer{
		check: c,
		ring:  ring,
		dot:   dot,
		label: label,
		objs:  []fyne.CanvasObject{ring, dot, label},
	}
}

func (c *customCheck) Tapped(_ *fyne.PointEvent) {
	c.Checked = !c.Checked
	if c.OnTappedFunc != nil {
		c.OnTappedFunc(c.Checked)
	}
	c.Refresh()
}

type customCheckRenderer struct {
	check *customCheck
	ring  *canvas.Circle
	dot   *canvas.Circle
	label *widget.Label
	objs  []fyne.CanvasObject
}

func (r *customCheckRenderer) Destroy()                     {}
func (r *customCheckRenderer) Objects() []fyne.CanvasObject { return r.objs }
func (r *customCheckRenderer) Refresh() {
	if r.check.Checked {
		r.dot.Show()
	} else {
		r.dot.Hide()
	}
	r.label.SetText(r.check.Label)
	r.ring.Refresh()
	r.label.Refresh()
}

func (r *customCheckRenderer) Layout(s fyne.Size) {
	iconSize := float32(16)
	padding := float32(8)

	r.ring.Resize(fyne.NewSize(iconSize, iconSize))
	r.ring.Move(fyne.NewPos(0, (s.Height-iconSize)/2))

	dotSize := iconSize - 6
	r.dot.Resize(fyne.NewSize(dotSize, dotSize))
	r.dot.Move(fyne.NewPos(3, (s.Height-iconSize)/2+3))

	labelMin := r.label.MinSize()
	labelW := s.Width - (iconSize + padding)
	if labelW < labelMin.Width {
		labelW = labelMin.Width
	}
	r.label.Resize(fyne.NewSize(labelW, s.Height))
	r.label.Move(fyne.NewPos(iconSize+padding, 0))
}

func (r *customCheckRenderer) MinSize() fyne.Size {
	l := r.label.MinSize()
	return fyne.NewSize(l.Width+24, l.Height)
}

func (m *MosuWidget) RefreshContent() {
	m.contentVBox.Objects = nil

	lines := strings.Split(m.rawText, "\n")
	for i, line := range lines {
		trimLine := strings.TrimSpace(line)
		lineIdx := i

		if strings.HasPrefix(trimLine, "[] ") || strings.HasPrefix(trimLine, "[ ] ") {
			label := strings.TrimPrefix(trimLine, "[] ")
			label = strings.TrimPrefix(label, "[ ] ")

			chk := newCustomCheck(label, false, func(b bool) {
				m.toggleLineState(lineIdx, b)
			})
			m.contentVBox.Add(chk)

		} else if strings.HasPrefix(trimLine, "[x] ") || strings.HasPrefix(trimLine, "[X] ") {
			label := strings.TrimPrefix(trimLine, "[x] ")
			label = strings.TrimPrefix(label, "[X] ")

			chk := newCustomCheck(label, true, func(b bool) {
				m.toggleLineState(lineIdx, b)
			})
			m.contentVBox.Add(chk)

		} else if strings.HasPrefix(trimLine, "- ") {
			labelTxt := strings.TrimPrefix(trimLine, "- ")
			label := widget.NewLabel("- " + labelTxt)
			label.Wrapping = fyne.TextWrapWord
			m.contentVBox.Add(label)
		} else {
			rt := widget.NewRichTextFromMarkdown(line)
			rt.Wrapping = fyne.TextWrapWord
			m.contentVBox.Add(rt)
		}
	}
	m.contentVBox.Refresh()
}

func (m *MosuWidget) toggleLineState(lineIdx int, checked bool) {
	lines := strings.Split(m.rawText, "\n")
	if lineIdx < 0 || lineIdx >= len(lines) {
		return
	}

	line := lines[lineIdx]
	trimLine := strings.TrimSpace(line)
	var content string

	if strings.HasPrefix(trimLine, "[] ") {
		content = strings.TrimPrefix(trimLine, "[] ")
	} else if strings.HasPrefix(trimLine, "[ ] ") {
		content = strings.TrimPrefix(trimLine, "[ ] ")
	} else if strings.HasPrefix(trimLine, "[x] ") {
		content = strings.TrimPrefix(trimLine, "[x] ")
	} else if strings.HasPrefix(trimLine, "[X] ") {
		content = strings.TrimPrefix(trimLine, "[X] ")
	} else {
		// Fallback for tight formats
		if strings.HasPrefix(trimLine, "[]") {
			content = strings.TrimPrefix(trimLine, "[]")
		}
		if strings.HasPrefix(trimLine, "[x]") {
			content = strings.TrimPrefix(trimLine, "[x]")
		}
	}

	newPrefix := "[] "
	if checked {
		newPrefix = "[x] "
	}
	lines[lineIdx] = newPrefix + content
	m.rawText = strings.Join(lines, "\n")
}

func (m *MosuWidget) DoubleTapped(_ *fyne.PointEvent) {
	m.ShowEditModal()
}

func (m *MosuWidget) Tapped(_ *fyne.PointEvent) {
}

// ShowEditModal enables in-place editing
func (m *MosuWidget) ShowEditModal() {
	if m.container == nil || len(m.container.Objects) < 2 {
		return
	}

	entry := widget.NewMultiLineEntry()
	entry.SetText(m.rawText)
	entry.Wrapping = fyne.TextWrapWord
	entry.SetPlaceHolder("Enter text...")

	entry.OnChanged = func(s string) {
		m.rawText = s
	}

	paddedEdit := container.NewPadded(entry)
	m.container.Objects = []fyne.CanvasObject{m.bg, paddedEdit}
	m.container.Refresh()

	c := fyne.CurrentApp().Driver().CanvasForObject(m)
	if c != nil {
		c.Focus(entry)
	}
}

// SetSelected toggles selection state and saves content if deselected
func (m *MosuWidget) SetSelected(selected bool) {
	if !selected && m.container != nil && len(m.container.Objects) > 1 {
		// Check if we are in edit mode to save and close
		if padded, ok := m.container.Objects[1].(*fyne.Container); ok && len(padded.Objects) > 0 {
			if _, isEntry := padded.Objects[0].(*widget.Entry); isEntry {
				m.RefreshContent()
				m.container.Objects = []fyne.CanvasObject{m.bg, container.NewPadded(m.contentVBox)}
				m.container.Refresh()
			}
		}
	}

	m.isSelected = selected
	if selected {
		m.bg.StrokeColor = theme.SelectionBlue
		m.bg.StrokeWidth = 2
	} else {
		m.bg.StrokeColor = color.Transparent
		m.bg.StrokeWidth = 0
	}
	m.bg.Refresh()
}

func (m *MosuWidget) FocusEntry() {
	m.ShowEditModal()
}

func (m *MosuWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(m.container)
}

func (m *MosuWidget) RefreshLayout(zoom float32) {
	width := m.WorldSize.Width * zoom
	height := m.WorldSize.Height * zoom
	x := m.WorldPos.X * zoom
	y := m.WorldPos.Y * zoom

	m.Move(fyne.NewPos(x, y))
	m.Resize(fyne.NewSize(width, height))
}
