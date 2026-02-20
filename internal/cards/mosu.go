package cards

import (
	"image/color"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"mosugo/internal/theme"
)

// coloredLabel creates a label with custom text color and wrapping
type coloredLabel struct {
	widget.Label
	TextColor color.Color
}

func newColoredLabel(text string, col color.Color) *coloredLabel {
	l := &coloredLabel{
		TextColor: col,
	}
	l.Text = text
	l.Wrapping = fyne.TextWrapWord
	l.ExtendBaseWidget(l)
	return l
}

func (l *coloredLabel) CreateRenderer() fyne.WidgetRenderer {
	l.ExtendBaseWidget(l)
	text := canvas.NewText(l.Text, l.TextColor)
	text.TextSize = 14
	text.Alignment = fyne.TextAlignLeading

	return &coloredLabelRenderer{
		label: l,
		text:  text,
		objs:  []fyne.CanvasObject{text},
	}
}

type coloredLabelRenderer struct {
	label *coloredLabel
	text  *canvas.Text
	objs  []fyne.CanvasObject
}

func (r *coloredLabelRenderer) Destroy() {}
func (r *coloredLabelRenderer) Objects() []fyne.CanvasObject {
	return r.objs
}

func (r *coloredLabelRenderer) Refresh() {
	r.text.Text = r.label.Text
	r.text.Color = r.label.TextColor
	r.text.Refresh()
}

func (r *coloredLabelRenderer) Layout(size fyne.Size) {
	// Handle text with existing newlines and wrap long lines
	if strings.Contains(r.label.Text, "\n") {
		// Text already has newlines, preserve them
		r.text.Text = r.label.Text
	} else {
		// Wrap text if it's too long
		wrapped := wrapText(r.label.Text, size.Width, 14)
		r.text.Text = wrapped
	}
	r.text.Resize(size)
}

func (r *coloredLabelRenderer) MinSize() fyne.Size {
	return r.text.MinSize()
}

func wrapText(text string, maxWidth float32, textSize float32) string {
	if maxWidth <= 0 || text == "" {
		return text
	}

	// Approximate character width for Comic Sans
	charWidth := textSize * 0.55
	maxChars := int(maxWidth / charWidth)
	if maxChars < 3 {
		maxChars = 3
	}

	// Handle very long words that exceed maxChars
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	currentLine := ""

	for _, word := range words {
		// If word itself is too long, break it
		if len(word) > maxChars {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = ""
			}
			// Break long word into chunks
			for len(word) > maxChars {
				lines = append(lines, word[:maxChars])
				word = word[maxChars:]
			}
			if len(word) > 0 {
				currentLine = word
			}
			continue
		}

		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= maxChars {
			currentLine = testLine
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

// paddedLayout adds uniform padding around content
type paddedLayout struct {
	padding float32
}

func (p *paddedLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	padding := p.padding
	pos := fyne.NewPos(padding, padding)
	contentSize := fyne.NewSize(
		size.Width-2*padding,
		size.Height-2*padding,
	)

	for _, obj := range objects {
		if obj.Visible() {
			obj.Move(pos)
			obj.Resize(contentSize)
		}
	}
}

func (p *paddedLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minSize := fyne.NewSize(0, 0)
	for _, obj := range objects {
		if obj.Visible() {
			objMin := obj.MinSize()
			if objMin.Width > minSize.Width {
				minSize.Width = objMin.Width
			}
			if objMin.Height > minSize.Height {
				minSize.Height = objMin.Height
			}
		}
	}
	padding := p.padding * 2
	return fyne.NewSize(minSize.Width+padding, minSize.Height+padding)
}

// compactVBoxLayout reduces spacing between elements
type compactVBoxLayout struct {
	spacing float32
}

func (c *compactVBoxLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	pos := fyne.NewPos(0, 0)
	for _, child := range objects {
		if !child.Visible() {
			continue
		}
		childSize := child.MinSize()
		child.Resize(fyne.NewSize(size.Width, childSize.Height))
		child.Move(pos)
		pos = pos.AddXY(0, childSize.Height+c.spacing)
	}
}

func (c *compactVBoxLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	minSize := fyne.NewSize(0, 0)
	for i, child := range objects {
		if !child.Visible() {
			continue
		}
		childMin := child.MinSize()
		minSize.Width = fyne.Max(minSize.Width, childMin.Width)
		minSize.Height += childMin.Height
		if i < len(objects)-1 {
			minSize.Height += c.spacing
		}
	}
	return minSize
}

// MosuWidget represents a card on the canvas
type MosuWidget struct {
	widget.BaseWidget
	ID        string
	CreatedAt time.Time
	ColorIndex int

	WorldPos  fyne.Position
	WorldSize fyne.Size

	bg          *canvas.Rectangle
	contentVBox *fyne.Container
	container   *fyne.Container

	isSelected bool
	rawText    string
	cursorPos  int
}

func NewMosuWidget(id string, c color.Color, colorIndex int) *MosuWidget {
	m := &MosuWidget{
		ID:         id,
		CreatedAt:  time.Now(),
		ColorIndex: colorIndex,
	}
	m.ExtendBaseWidget(m)

	m.bg = canvas.NewRectangle(c)
	m.bg.CornerRadius = 10
	m.bg.StrokeColor = color.Transparent
	m.bg.StrokeWidth = 0
	m.bg.FillColor = c

	m.rawText = ""
	m.cursorPos = 0
	m.contentVBox = container.NewVBox()
	m.contentVBox.Layout = &compactVBoxLayout{spacing: 2}

	// Add scroll container to handle vertical overflow
	scrollContent := container.NewVScroll(m.contentVBox)
	
	// Add custom padding around scrollable content
	paddedContent := container.New(
		&paddedLayout{padding: 16},
		scrollContent,
	)

	m.container = container.NewStack(m.bg, paddedContent)

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
	ring.StrokeColor = theme.InkWhite
	ring.StrokeWidth = 2

	dot := canvas.NewCircle(theme.InkWhite)
	if !c.Checked {
		dot.Hide()
	}

	// Use colored label for text wrapping
	label := newColoredLabel(c.Label, theme.InkWhite)

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
	label *coloredLabel
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
	r.label.Text = r.check.Label
	r.ring.Refresh()
	r.label.Refresh()
}

func (r *customCheckRenderer) Layout(s fyne.Size) {
	iconSize := float32(16)
	padding := float32(8)

	// CHECKBOX POSITION - Adjust these Y values to align with text
	// Current: Y=4 for ring, Y=7 for dot
	// Increase Y to move checkbox DOWN
	// Decrease Y to move checkbox UP
	r.ring.Resize(fyne.NewSize(iconSize, iconSize))
	r.ring.Move(fyne.NewPos(0, 4)) //

	dotSize := iconSize - 6
	r.dot.Resize(fyne.NewSize(dotSize, dotSize))
	r.dot.Move(fyne.NewPos(3, 7)) //

	// Label positioning with wrapping
	labelWidth := s.Width - (iconSize + padding)
	if labelWidth < 10 {
		labelWidth = 10
	}
	labelMinSize := r.label.MinSize()
	r.label.Resize(fyne.NewSize(labelWidth, labelMinSize.Height))
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
			label := newColoredLabel("• "+labelTxt, theme.InkWhite)
			m.contentVBox.Add(label)
		} else if trimLine != "" {
			// Regular text
			label := newColoredLabel(line, theme.InkWhite)
			m.contentVBox.Add(label)
		} else {
			// Empty line - add small spacer
			spacer := newColoredLabel(" ", theme.InkWhite)
			m.contentVBox.Add(spacer)
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

func (m *MosuWidget) Tapped(_ *fyne.PointEvent) {
	// Focus this card for keyboard input
	c := fyne.CurrentApp().Driver().CanvasForObject(m)
	if c != nil {
		c.Focus(m)
	}
}

func (m *MosuWidget) FocusGained() {
	
}

func (m *MosuWidget) FocusLost() {

	m.RefreshContent()
}

func (m *MosuWidget) TypedRune(r rune) {

	if m.cursorPos < 0 {
		m.cursorPos = 0
	}
	if m.cursorPos > len(m.rawText) {
		m.cursorPos = len(m.rawText)
	}

	m.rawText = m.rawText[:m.cursorPos] + string(r) + m.rawText[m.cursorPos:]
	m.cursorPos++

	m.RefreshContent()
}

func (m *MosuWidget) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyBackspace:
		if m.cursorPos > 0 {
			m.rawText = m.rawText[:m.cursorPos-1] + m.rawText[m.cursorPos:]
			m.cursorPos--
			m.RefreshContent()
		}
	case fyne.KeyDelete:
		if m.cursorPos < len(m.rawText) {
			m.rawText = m.rawText[:m.cursorPos] + m.rawText[m.cursorPos+1:]
			m.RefreshContent()
		}
	case fyne.KeyReturn:
		m.rawText = m.rawText[:m.cursorPos] + "\n" + m.rawText[m.cursorPos:]
		m.cursorPos++
		m.RefreshContent()
	case fyne.KeyLeft:
		if m.cursorPos > 0 {
			m.cursorPos--
		}
	case fyne.KeyRight:
		if m.cursorPos < len(m.rawText) {
			m.cursorPos++
		}
	case fyne.KeyUp, fyne.KeyDown, fyne.KeyHome, fyne.KeyEnd:

	}
}

// SetSelected toggles selection state
func (m *MosuWidget) SetSelected(selected bool) {
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

// GetText returns the raw text content of the card
func (m *MosuWidget) GetText() string {
	return m.rawText
}

// SetText sets the text content of the card and refreshes the display
func (m *MosuWidget) SetText(text string) {
	m.rawText = text
	m.cursorPos = len(text)
}
