package canvas

import (
	"image/color"

	"fyne.io/fyne/v2/canvas"
)

func BoxGridPattern(c *MosugoCanvas, gridSize int, lineColor, bgColor color.Color) *canvas.Raster {
	return canvas.NewRasterWithPixels(func(x, y, w, h int) color.Color {
		wx := x - int(c.Offset.X)
		wy := y - int(c.Offset.Y)

		if wx%gridSize == 0 && (wy/4)%2 == 0 {
			return lineColor
		}
		if wy%gridSize == 0 && (wx/4)%2 == 0 {
			return lineColor
		}

		return bgColor
	})
}
