package canvas

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2/canvas"
)

func BoxGridPattern(c *MosugoCanvas, gridSize int, lineColor, bgColor color.Color) *canvas.Raster {
	return canvas.NewRasterWithPixels(func(x, y, w, h int) color.Color {
		devScale := float64(c.DeviceScale)
		if devScale <= 0 {
			devScale = 1.0
		}

		gSize := float64(gridSize) * devScale

		distX := math.Abs(math.Remainder(float64(x), gSize))
		distY := math.Abs(math.Remainder(float64(y), gSize))

		thickness := 1 * devScale

		isVertical := distX < thickness
		isHorizontal := distY < thickness

		if !isVertical && !isHorizontal {
			return bgColor
		}

		wx := (float64(x) / devScale)
		wy := (float64(y) / devScale)

		if isVertical {
			if int(math.Abs(wy)/3)%2 == 0 {
				return lineColor
			}
		}
		if isHorizontal {
			if int(math.Abs(wx)/3)%2 == 0 {
				return lineColor
			}
		}

		return bgColor
	})
}
