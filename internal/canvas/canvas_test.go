package canvas

import (
	"testing"

	"mosugo/internal/testutil"

	"fyne.io/fyne/v2"
	"github.com/stretchr/testify/assert"
)

// TestScreenToWorldBasic tests basic coordinate transformations at 1:1 scale with zero offset
func TestScreenToWorldBasic(t *testing.T) {
	c := NewMosugoCanvas()
	c.Scale = 1.0
	c.Offset = fyne.NewPos(0, 0)

	tests := []struct {
		name     string
		screen   fyne.Position
		expected fyne.Position
	}{
		{"Origin", fyne.NewPos(0, 0), fyne.NewPos(0, 0)},
		{"Positive", fyne.NewPos(100, 200), fyne.NewPos(100, 200)},
		{"Negative", fyne.NewPos(-50, -100), fyne.NewPos(-50, -100)},
		{"Fractional", fyne.NewPos(10.5, 20.3), fyne.NewPos(10.5, 20.3)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.ScreenToWorld(tt.screen)
			testutil.PositionEqual(t, tt.expected, result)
		})
	}
}

// TestScreenToWorldAtScale tests coordinate transformations at various zoom levels
func TestScreenToWorldAtScale(t *testing.T) {
	tests := []struct {
		name     string
		scale    float32
		offset   fyne.Position
		screen   fyne.Position
		expected fyne.Position
	}{
		{"Scale 0.1 origin", 0.1, fyne.NewPos(0, 0), fyne.NewPos(10, 10), fyne.NewPos(100, 100)},
		{"Scale 0.5", 0.5, fyne.NewPos(0, 0), fyne.NewPos(50, 100), fyne.NewPos(100, 200)},
		{"Scale 2.0", 2.0, fyne.NewPos(0, 0), fyne.NewPos(200, 100), fyne.NewPos(100, 50)},
		{"Scale 5.0", 5.0, fyne.NewPos(0, 0), fyne.NewPos(500, 250), fyne.NewPos(100, 50)},
		{"Scale 10.0", 10.0, fyne.NewPos(0, 0), fyne.NewPos(1000, 500), fyne.NewPos(100, 50)},
		{"Scale 1.0 with offset", 1.0, fyne.NewPos(100, 50), fyne.NewPos(200, 150), fyne.NewPos(100, 100)},
		{"Scale 2.0 with offset", 2.0, fyne.NewPos(100, 50), fyne.NewPos(300, 250), fyne.NewPos(100, 100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewMosugoCanvas()
			c.Scale = tt.scale
			c.Offset = tt.offset

			result := c.ScreenToWorld(tt.screen)
			testutil.PositionEqual(t, tt.expected, result)
		})
	}
}

// TestWorldToScreenBasic tests basic world-to-screen transformations
func TestWorldToScreenBasic(t *testing.T) {
	c := NewMosugoCanvas()
	c.Scale = 1.0
	c.Offset = fyne.NewPos(0, 0)

	tests := []struct {
		name     string
		world    fyne.Position
		expected fyne.Position
	}{
		{"Origin", fyne.NewPos(0, 0), fyne.NewPos(0, 0)},
		{"Positive", fyne.NewPos(100, 200), fyne.NewPos(100, 200)},
		{"Negative", fyne.NewPos(-50, -100), fyne.NewPos(-50, -100)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.WorldToScreen(tt.world)
			testutil.PositionEqual(t, tt.expected, result)
		})
	}
}

// TestWorldToScreenAtScale tests world-to-screen transformations at various zoom levels
func TestWorldToScreenAtScale(t *testing.T) {
	tests := []struct {
		name     string
		scale    float32
		offset   fyne.Position
		world    fyne.Position
		expected fyne.Position
	}{
		{"Scale 0.1", 0.1, fyne.NewPos(0, 0), fyne.NewPos(100, 100), fyne.NewPos(10, 10)},
		{"Scale 0.5", 0.5, fyne.NewPos(0, 0), fyne.NewPos(100, 200), fyne.NewPos(50, 100)},
		{"Scale 2.0", 2.0, fyne.NewPos(0, 0), fyne.NewPos(100, 50), fyne.NewPos(200, 100)},
		{"Scale 5.0", 5.0, fyne.NewPos(0, 0), fyne.NewPos(100, 50), fyne.NewPos(500, 250)},
		{"Scale 10.0", 10.0, fyne.NewPos(0, 0), fyne.NewPos(100, 50), fyne.NewPos(1000, 500)},
		{"Scale 1.0 with offset", 1.0, fyne.NewPos(100, 50), fyne.NewPos(100, 100), fyne.NewPos(200, 150)},
		{"Scale 2.0 with offset", 2.0, fyne.NewPos(100, 50), fyne.NewPos(100, 100), fyne.NewPos(300, 250)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewMosugoCanvas()
			c.Scale = tt.scale
			c.Offset = tt.offset

			result := c.WorldToScreen(tt.world)
			testutil.PositionEqual(t, tt.expected, result)
		})
	}
}

// TestCoordinateTransformRoundtrip verifies bidirectional conversion accuracy
func TestCoordinateTransformRoundtrip(t *testing.T) {
	scales := testutil.SampleScales()
	positions := testutil.SamplePositions()

	for _, scale := range scales {
		t.Run("Scale_"+formatFloat(scale), func(t *testing.T) {
			c := NewMosugoCanvas()
			c.Scale = scale
			c.Offset = fyne.NewPos(50, 100) // Non-zero offset for comprehensive test

			for i, pos := range positions {
				t.Run("Position_"+string(rune(i)), func(t *testing.T) {
					// World -> Screen -> World
					screen := c.WorldToScreen(pos)
					backToWorld := c.ScreenToWorld(screen)
					testutil.PositionEqual(t, pos, backToWorld)

					// Screen -> World -> Screen
					world := c.ScreenToWorld(pos)
					backToScreen := c.WorldToScreen(world)
					testutil.PositionEqual(t, pos, backToScreen)
				})
			}
		})
	}
}

// TestCoordinateTransformEdgeCases tests edge cases and extreme values
func TestCoordinateTransformEdgeCases(t *testing.T) {
	t.Run("Very large coordinates", func(t *testing.T) {
		c := NewMosugoCanvas()
		c.Scale = 1.0
		c.Offset = fyne.NewPos(0, 0)

		large := fyne.NewPos(1e6, 1e6)
		screen := c.WorldToScreen(large)
		backToWorld := c.ScreenToWorld(screen)

		// Allow slightly larger tolerance for very large numbers
		assert.InDelta(t, large.X, backToWorld.X, 0.1)
		assert.InDelta(t, large.Y, backToWorld.Y, 0.1)
	})

	t.Run("Very small scale", func(t *testing.T) {
		c := NewMosugoCanvas()
		c.Scale = 0.01
		c.Offset = fyne.NewPos(0, 0)

		world := fyne.NewPos(1000, 1000)
		screen := c.WorldToScreen(world)
		assert.InDelta(t, 10.0, screen.X, testutil.PositionTolerance)
		assert.InDelta(t, 10.0, screen.Y, testutil.PositionTolerance)
	})

	t.Run("Negative scale should work", func(t *testing.T) {
		c := NewMosugoCanvas()
		c.Scale = -1.0 // Mathematically valid, creates mirror
		c.Offset = fyne.NewPos(0, 0)

		world := fyne.NewPos(100, 100)
		screen := c.WorldToScreen(world)
		assert.Equal(t, float32(-100), screen.X)
		assert.Equal(t, float32(-100), screen.Y)
	})
}

// TestSnapToGrid tests grid snapping at GridSize intervals
func TestSnapToGrid(t *testing.T) {
	c := NewMosugoCanvas()

	tests := []struct {
		name     string
		input    float32
		expected float32
	}{
		{"Zero", 0, 0},
		{"Exact grid", 30, 30},
		{"Between grids low", 14, 0},
		{"Between grids mid", 15, 0},
		{"Between grids high", 29, 0},
		{"Next grid", 31, 30},
		{"Multiple grids", 60, 60},
		{"Large value", 90, 90},
		{"Negative at boundary", -30, -30},
		{"Negative between", -15, -30},
		{"Negative low", -31, -60},
		{"Fractional positive", 14.9, 0},
		{"Fractional negative", -14.9, -30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.Snap(tt.input)
			testutil.Float32Equal(t, tt.expected, result)
		})
	}
}

// TestSnapUpToGrid tests upward grid snapping
func TestSnapUpToGrid(t *testing.T) {
	c := NewMosugoCanvas()

	tests := []struct {
		name     string
		input    float32
		expected float32
	}{
		{"Zero", 0, 0},
		{"Exact grid", 30, 30},
		{"Between grids low", 1, 30},
		{"Between grids mid", 15, 30},
		{"Between grids high", 29, 30},
		{"Next grid", 31, 60},
		{"Multiple grids", 60, 60},
		{"Negative at boundary", -30, -30},
		{"Negative between", -29, 0},
		{"Negative low", -31, -30},
		{"Fractional positive", 0.1, 30},
		{"Fractional at grid", 30.0, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.SnapUp(tt.input)
			testutil.Float32Equal(t, tt.expected, result)
		})
	}
}

// TestGetSetOffset tests offset getters and setters
func TestGetSetOffset(t *testing.T) {
	c := NewMosugoCanvas()

	offset := fyne.NewPos(123.45, 678.90)
	c.SetOffset(offset)
	result := c.GetOffset()

	testutil.PositionEqual(t, offset, result)
}

// TestGetScale tests scale getter
func TestGetScale(t *testing.T) {
	c := NewMosugoCanvas()
	c.Scale = 2.5

	result := c.GetScale()
	testutil.Float32Equal(t, 2.5, result)
}

// TestNewMosugoCanvasDefaults tests that canvas is initialized with correct defaults
func TestNewMosugoCanvasDefaults(t *testing.T) {
	c := NewMosugoCanvas()

	assert.NotNil(t, c)
	testutil.Float32Equal(t, 1.0, c.Scale)
	testutil.PositionEqual(t, fyne.NewPos(0, 0), c.Offset)
	assert.NotNil(t, c.Content)
	assert.NotNil(t, c.strokesMap)
	assert.NotNil(t, c.strokeIDMap)
	assert.NotNil(t, c.glowLines)
	assert.Equal(t, 1, c.nextStrokeID)
	assert.False(t, c.isDirty)
}

// Helper function to format float for test names
func formatFloat(f float32) string {
	if f == float32(int(f)) {
		return string(rune(int(f) + '0'))
	}
	// For non-integers, use a simplified representation
	return string(rune(int(f*10) + '0'))
}
