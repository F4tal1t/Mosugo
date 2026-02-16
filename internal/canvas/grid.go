package canvas

import (
	"image/color"

	"fyne.io/fyne/v2/canvas"
)

// Now accepts offset and zoom to render dynamic grid
func DotGridPattern(c *MosugoCanvas, gridSize int, dotColor, bgColor color.Color) *canvas.Raster {
	return canvas.NewRasterWithPixels(func(x, y, w, h int) color.Color {
		// Adjust coordinates for Zoom and Offset
		// Screen(x) -> World(wx) = (x - Offset.X) / Zoom

		zoomedSize := float32(gridSize) * c.Zoom
		if zoomedSize < 5 {
			zoomedSize = 5
		} // prevent div by zero or tiny grid

		// Calculate relative position within a grid cell
		// To make grid move with Offset:
		// We use (x - Offset) modulo (GridSize * Zoom)

		offX := float32(x) - c.Offset.X
		offY := float32(y) - c.Offset.Y

		// Standard modulo logic for float
		iz := int(zoomedSize)
		modX := int(offX) % iz
		modY := int(offY) % iz

		// Handle negative modulo correctly
		if modX < 0 {
			modX += iz
		}
		if modY < 0 {
			modY += iz
		}

		if modX < 4 && modY < 4 { // Slightly larger dots for visibility
			return dotColor
		}
		return bgColor
	})
}
