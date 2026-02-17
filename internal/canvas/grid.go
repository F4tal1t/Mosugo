package canvas

import (
	"image/color"
	"math"

	"fyne.io/fyne/v2/canvas"
)

func BoxGridPattern(c *MosugoCanvas, gridSize int, lineColor, bgColor color.Color) *canvas.Raster {
	return canvas.NewRasterWithPixels(func(x, y, w, h int) color.Color {
		zoom := float64(c.Scale)
		if zoom <= 0 {
			zoom = 1.0
		}

		devScale := float64(c.DeviceScale)
		if devScale <= 0 {
			devScale = 1.0
		}

		offX := float64(c.Offset.X)
		offY := float64(c.Offset.Y)
		gSize := float64(gridSize)

		lx := (float64(x) + 0.5) / devScale
		ly := (float64(y) + 0.5) / devScale

		wx := (lx - offX) / zoom
		wy := (ly - offY) / zoom

		distX := math.Abs(math.Remainder(wx, gSize))
		distY := math.Abs(math.Remainder(wy, gSize))

		physDistX := distX * zoom * devScale
		physDistY := distY * zoom * devScale

		const thickness = 0.55

		isVertical := physDistX < thickness
		isHorizontal := physDistY < thickness

		if !isVertical && !isHorizontal {
			return bgColor
		}

		if isVertical {
			if int(math.Abs(wy)/4)%2 == 0 {
				return lineColor
			}
		}
		if isHorizontal {
			if int(math.Abs(wx)/4)%2 == 0 {
				return lineColor
			}
		}

		return bgColor
	})
}
