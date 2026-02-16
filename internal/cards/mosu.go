package cards

import (
	"image/color"

	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type MosuWidget struct {
	widget.BaseWidget
	ID string

	// World Coordinates (Unzoomed)
	WorldPos  fyne.Position
	WorldSize fyne.Size

	// Components
	bg      *canvas.Rectangle
	content *widget.Entry // Multiline text entry
}

func NewMosuWidget(id string, c color.Color) *MosuWidget {
	m := &MosuWidget{ID: id}
	m.ExtendBaseWidget(m)

	m.bg = canvas.NewRectangle(c)
	m.bg.CornerRadius = 8
	m.bg.StrokeColor = color.Black
	m.bg.StrokeWidth = 1

	m.content = widget.NewMultiLineEntry()
	m.content.Wrapping = fyne.TextWrapWord
	m.content.SetPlaceHolder("Type your note...")

	// Simple Checklist Logic: Replace "[] " with "☐ " while typing
	m.content.OnChanged = func(s string) {
		start := m.content.CursorColumn
		row := m.content.CursorRow

		if strings.Contains(s, "[] ") {
			newText := strings.ReplaceAll(s, "[] ", "☐ ")
			if newText != s {
				m.content.SetText(newText)
				// Attempt to restore cursor (simple approximation)
				m.content.CursorColumn = start - 2 // removed 3 chars "[] ", added 1 "☐ "
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

	// Remove default Entry border for cleaner look
	m.content.Resize(fyne.NewSize(100, 100))

	return m
}

func (m *MosuWidget) CreateRenderer() fyne.WidgetRenderer {
	// Padded container: background + content with margin
	c := container.NewStack(
		m.bg,
		container.NewPadded(m.content),
	)
	return widget.NewSimpleRenderer(c)
}

func (m *MosuWidget) RefreshLayout(zoom float32) {
	// Simple Scale Logic: World * Zoom
	width := m.WorldSize.Width * zoom
	height := m.WorldSize.Height * zoom

	x := m.WorldPos.X * zoom
	y := m.WorldPos.Y * zoom

	m.Move(fyne.NewPos(x, y))
	m.Resize(fyne.NewSize(width, height))
}
