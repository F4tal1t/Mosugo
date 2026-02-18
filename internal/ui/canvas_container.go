package ui

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type MetaballBorder struct {
	widget.BaseWidget
	BorderColor color.Color
	Thickness   float32
	TabY        float32
	TabWidth    float32
	TabHeight   float32
}

func NewMetaballBorder(c color.Color) *MetaballBorder {
	m := &MetaballBorder{
		BorderColor: c,
		Thickness:   10,
		TabY:        100,
		TabWidth:    60,
		TabHeight:   180,
	}
	m.ExtendBaseWidget(m)
	return m
}

func (m *MetaballBorder) CreateRenderer() fyne.WidgetRenderer {
	r := &metaballRenderer{m: m}
	r.raster = canvas.NewRaster(r.generator)
	return r
}

type metaballRenderer struct {
	m      *MetaballBorder
	raster *canvas.Raster
}

func (r *metaballRenderer) MinSize() fyne.Size {
	return fyne.NewSize(100, 100)
}

func (r *metaballRenderer) Layout(s fyne.Size) {
	r.raster.Resize(s)
}

func (r *metaballRenderer) Refresh() {
	r.raster.Refresh()
}

func (r *metaballRenderer) Destroy() {}

func (r *metaballRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.raster}
}

func sdBox(pX, pY, bX, bY float32) float32 {
	dX := float32(math.Abs(float64(pX))) - bX
	dY := float32(math.Abs(float64(pY))) - bY
	mX := float32(math.Max(float64(dX), 0))
	mY := float32(math.Max(float64(dY), 0))
	distOutside := float32(math.Sqrt(float64(mX*mX + mY*mY)))
	distInside := float32(math.Min(math.Max(float64(dX), float64(dY)), 0))
	return distOutside + distInside
}

func opSmoothUnion(d1, d2, k float32) float32 {
	h := float32(math.Max(float64(k)-math.Abs(float64(d1-d2)), 0.0)) / k
	return float32(math.Min(float64(d1), float64(d2))) - h*h*k*0.25
}

func (r *metaballRenderer) generator(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	col := r.m.BorderColor
	rInt, gInt, bInt, aInt := col.RGBA()

	scale := float32(1.0)
	sz := r.m.Size()
	if sz.Width > 0 && w > 0 {
		scale = float32(w) / sz.Width
	}

	thick := r.m.Thickness * scale
	// tY := r.m.TabY * scale
	tW := r.m.TabWidth * scale
	tH := r.m.TabHeight * scale
	fW, fH := float32(w), float32(h)

	// Create hole coordinates (inner rectangle)
	holeW := fW - 2*thick
	holeH := fH - 2*thick
	holeHalfW := holeW / 2
	holeHalfH := holeH / 2
	centerX := fW / 2
	centerY := fH / 2

	// Normalized Color components
	fR := float32(rInt) / 65535.0
	fG := float32(gInt) / 65535.0
	fB := float32(bInt) / 65535.0
	fA := float32(aInt) / 65535.0

	k := 20 * scale // smoothness factor

	// Tab Position
	// Centered on the left inner border edge
	tabCX := -holeHalfW
	tabCY := float32(0)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			pX := float32(x)
			pY := float32(y)

			// 1. SDF for Frame
			relX := pX - centerX
			relY := pY - centerY

			rad := 12 * scale
			dRectHole := sdBox(relX, relY, holeHalfW-rad, holeHalfH-rad) - rad
			sdFrame := -dRectHole // Negative when Outside the Hole (Solid Frame)

			// 2. SDF for Tab
			tabRad := 15 * scale
			sdTab := sdBox(relX-tabCX, relY-tabCY, tW/1.4-tabRad, tH/2.8-tabRad) - tabRad

			// 3. Union (Merge Border and Tab)
			dFinal := opSmoothUnion(sdFrame, sdTab, k)

			// 4. Render
			alpha := float32(0.0)
			val := dFinal

			if val < -0.5 {
				alpha = 1.0
			} else if val > 0.5 {
				alpha = 0.0
			} else {
				t := (val + 0.5)
				alpha = 1.0 - t
			}

			if alpha > 0 {
				offset := (y*w + x) * 4
				img.Pix[offset] = uint8(fR * alpha * 255)
				img.Pix[offset+1] = uint8(fG * alpha * 255)
				img.Pix[offset+2] = uint8(fB * alpha * 255)
				img.Pix[offset+3] = uint8(fA * alpha * 255)
			}
		}
	}
	return img
}
