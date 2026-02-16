package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

type MosugoTheme struct {
	fyne.Theme
}

func NewMosugoTheme() fyne.Theme {
	return &MosugoTheme{Theme: theme.DefaultTheme()}
}

func (t *MosugoTheme) Font(s fyne.TextStyle) fyne.Resource {
	// Logic to load custom font:
	// if customFontLoaded { return customFont }
	// For now, return default
	return t.Theme.Font(s)
}

func (t *MosugoTheme) Color(n fyne.ThemeColorName, v fyne.ThemeVariant) color.Color {
	// Custom colors if needed
	return t.Theme.Color(n, v)
}
