package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type MosugoTheme struct {
	fyne.Theme
}

var (
	InkGrey       = color.RGBA{130, 130, 140, 255}
	InkLightGrey  = color.RGBA{160, 160, 170, 255}
	CardWhite     = color.RGBA{255, 255, 255, 255}
	CardYellow    = color.RGBA{255, 247, 164, 255}
	CardTurquoise = color.RGBA{160, 194, 255, 255}
	CardPink      = color.RGBA{255, 208, 196, 255}
	CardBg        = CardYellow // Default card background
	GridLine      = color.RGBA{190, 190, 190, 255}
	GridBg        = color.RGBA{220, 220, 220, 255}
	SelectionBlue = color.RGBA{100, 150, 255, 255}
)

func NewMosugoTheme() fyne.Theme {
	return &MosugoTheme{Theme: theme.DefaultTheme()}
}

func (t *MosugoTheme) Font(s fyne.TextStyle) fyne.Resource {
	return t.Theme.Font(s)
}

func (t *MosugoTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	switch n {
	case theme.ColorNameForeground:
		return InkGrey
	case theme.ColorNameBackground:
		return GridBg
	case theme.ColorNameButton:
		return color.RGBA{245, 245, 248, 255}
	case theme.ColorNameHover:
		return color.RGBA{0, 64, 92, 255}
	case theme.ColorNamePrimary:
		return color.RGBA{0, 113, 162, 255}
	case theme.ColorNameInputBackground:
		return color.Transparent
	}
	return t.Theme.Color(n, v)
}
