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
	InkGrey       = color.RGBA{80, 80, 90, 255}
	InkLightGrey  = color.RGBA{120, 120, 130, 255}
	CardWhite     = color.RGBA{255, 255, 255, 255}
	CardYellow    = color.RGBA{255, 247, 64, 255}
	CardTurquoise = color.RGBA{64, 224, 208, 255}
	CardPink      = color.RGBA{255, 105, 180, 255}
	GridLine      = color.RGBA{200, 200, 220, 255}
	GridBg        = color.RGBA{250, 250, 250, 255}
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
		return color.RGBA{230, 230, 235, 255}
	case theme.ColorNamePrimary:
		return SelectionBlue
	}
	return t.Theme.Color(n, v)
}
