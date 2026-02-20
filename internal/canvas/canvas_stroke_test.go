package canvas

import (
	"testing"

	"mosugo/internal/testutil"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSimplifyStrokeStraightLine tests that a straight line reduces to 2 points
func TestSimplifyStrokeStraightLine(t *testing.T) {
	c := NewMosugoCanvas()
	samples := testutil.SampleStrokePoints()

	result := c.SimplifyStroke(samples.StraightLine, 3.0)

	require.Len(t, result, 2, "Straight line should simplify to 2 points")
	testutil.PositionEqual(t, samples.StraightLine[0], result[0])
	testutil.PositionEqual(t, samples.StraightLine[len(samples.StraightLine)-1], result[1])
}

// TestSimplifyStrokeZigzag tests that zigzag pattern retains key points
func TestSimplifyStrokeZigzag(t *testing.T) {
	c := NewMosugoCanvas()
	samples := testutil.SampleStrokePoints()

	// Use higher epsilon to ensure simplification happens
	result := c.SimplifyStroke(samples.Zigzag, 5.0)

	// Zigzag should keep some intermediate points
	assert.Greater(t, len(result), 2, "Zigzag should keep more than 2 points")
	assert.Less(t, len(result), len(samples.Zigzag), "Should reduce point count")
}

// TestSimplifyStrokeCircle tests circle approximation
func TestSimplifyStrokeCircle(t *testing.T) {
	c := NewMosugoCanvas()
	samples := testutil.SampleStrokePoints()

	result := c.SimplifyStroke(samples.Circle, 3.0)

	// Circle with 100 points should be significantly reduced
	assert.Greater(t, len(result), 4, "Circle needs more than 4 points")
	assert.Less(t, len(result), 50, "Should reduce from 100 to less than 50 points")
}

// TestSimplifyStrokeEpsilonVariations tests different epsilon values
func TestSimplifyStrokeEpsilonVariations(t *testing.T) {
	c := NewMosugoCanvas()
	samples := testutil.SampleStrokePoints()

	tests := []struct {
		name            string
		epsilon         float32
		maxExpectedSize int
	}{
		{"Low epsilon 0.5", 0.5, 80},    // Less aggressive
		{"Medium epsilon 3.0", 3.0, 30}, // Balanced
		{"High epsilon 10.0", 10.0, 15}, // Very aggressive
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.SimplifyStroke(samples.Circle, tt.epsilon)
			assert.LessOrEqual(t, len(result), tt.maxExpectedSize,
				"Epsilon %.1f should reduce to max %d points", tt.epsilon, tt.maxExpectedSize)
		})
	}
}

// TestSimplifyStrokeEdgeCases tests edge cases for stroke simplification
func TestSimplifyStrokeEdgeCases(t *testing.T) {
	c := NewMosugoCanvas()
	samples := testutil.SampleStrokePoints()

	t.Run("Empty stroke", func(t *testing.T) {
		result := c.SimplifyStroke([]fyne.Position{}, 3.0)
		assert.Empty(t, result)
	})

	t.Run("Single point", func(t *testing.T) {
		result := c.SimplifyStroke(samples.SinglePoint, 3.0)
		require.Len(t, result, 1)
		testutil.PositionEqual(t, samples.SinglePoint[0], result[0])
	})

	t.Run("Two points", func(t *testing.T) {
		result := c.SimplifyStroke(samples.TwoPoints, 3.0)
		require.Len(t, result, 2)
		testutil.PositionEqual(t, samples.TwoPoints[0], result[0])
		testutil.PositionEqual(t, samples.TwoPoints[1], result[1])
	})

	t.Run("Collinear points", func(t *testing.T) {
		collinear := []fyne.Position{
			{X: 0, Y: 0},
			{X: 10, Y: 10},
			{X: 20, Y: 20},
			{X: 30, Y: 30},
		}
		result := c.SimplifyStroke(collinear, 1.0)
		require.Len(t, result, 2, "Collinear points should reduce to start and end")
	})
}

// TestGenerateStrokeID tests sequential stroke ID generation
func TestGenerateStrokeID(t *testing.T) {
	c := NewMosugoCanvas()

	// Should start at 1 (from initialization)
	id1 := c.GenerateStrokeID()
	assert.Equal(t, 1, id1)

	// Should increment sequentially
	id2 := c.GenerateStrokeID()
	assert.Equal(t, 2, id2)

	id3 := c.GenerateStrokeID()
	assert.Equal(t, 3, id3)
}

// TestGenerateStrokeIDNoCollisions tests that IDs don't collide
func TestGenerateStrokeIDNoCollisions(t *testing.T) {
	c := NewMosugoCanvas()
	ids := make(map[int]bool)

	// Generate 100 IDs
	for i := 0; i < 100; i++ {
		id := c.GenerateStrokeID()
		assert.False(t, ids[id], "ID %d already generated (collision)", id)
		ids[id] = true
	}

	assert.Len(t, ids, 100, "Should have 100 unique IDs")
}

// TestValidateStrokeID tests stroke ID validation
func TestValidateStrokeID(t *testing.T) {
	c := NewMosugoCanvas()

	tests := []struct {
		name     string
		id       int
		expected bool
	}{
		{"Zero is invalid", 0, false},
		{"Negative is invalid", -1, false},
		{"One is valid", 1, true},
		{"Large positive valid", 999, true},
		{"Very large valid", 1000000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := c.ValidateStrokeID(tt.id)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRegisterStrokeAndRetrieve tests stroke registration and retrieval
func TestRegisterStrokeAndRetrieve(t *testing.T) {
	c := NewMosugoCanvas()

	line := canvas.NewLine(nil)
	p1 := fyne.NewPos(10, 20)
	p2 := fyne.NewPos(100, 200)
	strokeID := 42

	// Register stroke
	c.RegisterStroke(line, p1, p2, strokeID)

	// Retrieve stroke ID
	retrievedID, ok := c.GetStrokeID(line)
	require.True(t, ok, "Should find registered stroke")
	assert.Equal(t, strokeID, retrievedID)

	// Retrieve stroke points
	rp1, rp2, ok := c.GetStrokePoints(line)
	require.True(t, ok, "Should find registered stroke points")
	testutil.PositionEqual(t, p1, rp1)
	testutil.PositionEqual(t, p2, rp2)
}

// TestGetStrokeIDNotFound tests retrieval of unregistered stroke
func TestGetStrokeIDNotFound(t *testing.T) {
	c := NewMosugoCanvas()

	line := canvas.NewLine(nil)
	id, ok := c.GetStrokeID(line)

	assert.False(t, ok, "Should not find unregistered stroke")
	assert.Equal(t, 0, id, "Should return zero for not found")
}

// TestGetStrokePointsNotFound tests retrieval of unregistered stroke points
func TestGetStrokePointsNotFound(t *testing.T) {
	c := NewMosugoCanvas()

	line := canvas.NewLine(nil)
	p1, p2, ok := c.GetStrokePoints(line)

	assert.False(t, ok, "Should not find unregistered stroke")
	testutil.PositionEqual(t, fyne.Position{}, p1)
	testutil.PositionEqual(t, fyne.Position{}, p2)
}

// TestRegisterMultipleStrokes tests registering multiple strokes
func TestRegisterMultipleStrokes(t *testing.T) {
	c := NewMosugoCanvas()

	strokes := []struct {
		line     *canvas.Line
		p1       fyne.Position
		p2       fyne.Position
		strokeID int
	}{
		{canvas.NewLine(nil), fyne.NewPos(0, 0), fyne.NewPos(10, 10), 1},
		{canvas.NewLine(nil), fyne.NewPos(20, 20), fyne.NewPos(30, 30), 2},
		{canvas.NewLine(nil), fyne.NewPos(40, 40), fyne.NewPos(50, 50), 3},
	}

	// Register all strokes
	for _, s := range strokes {
		c.RegisterStroke(s.line, s.p1, s.p2, s.strokeID)
	}

	// Verify all can be retrieved
	for i, s := range strokes {
		t.Run("Stroke_"+string(rune(i+'0')), func(t *testing.T) {
			id, ok := c.GetStrokeID(s.line)
			require.True(t, ok)
			assert.Equal(t, s.strokeID, id)

			p1, p2, ok := c.GetStrokePoints(s.line)
			require.True(t, ok)
			testutil.PositionEqual(t, s.p1, p1)
			testutil.PositionEqual(t, s.p2, p2)
		})
	}
}

// TestIsGlowLine tests glow line detection
func TestIsGlowLine(t *testing.T) {
	c := NewMosugoCanvas()

	glowLine := canvas.NewLine(nil)
	regularLine := canvas.NewLine(nil)

	// Mark one as glow
	c.glowLines[glowLine] = true

	assert.True(t, c.IsGlowLine(glowLine), "Should detect glow line")
	assert.False(t, c.IsGlowLine(regularLine), "Should not detect regular line as glow")
}

// TestPerpendicularDistance tests the internal distance calculation
func TestPerpendicularDistance(t *testing.T) {
	tests := []struct {
		name     string
		point    fyne.Position
		start    fyne.Position
		end      fyne.Position
		expected float32
	}{
		{
			name:     "Point on line",
			point:    fyne.NewPos(50, 50),
			start:    fyne.NewPos(0, 0),
			end:      fyne.NewPos(100, 100),
			expected: 0,
		},
		{
			name:     "Point perpendicular to middle",
			point:    fyne.NewPos(60, 40),
			start:    fyne.NewPos(0, 0),
			end:      fyne.NewPos(100, 100),
			expected: 14.142, // sqrt(200) ≈ 14.142
		},
		{
			name:     "Point beyond start",
			point:    fyne.NewPos(-10, -10),
			start:    fyne.NewPos(0, 0),
			end:      fyne.NewPos(100, 100),
			expected: 14.142, // Distance to start point
		},
		{
			name:     "Point beyond end",
			point:    fyne.NewPos(110, 110),
			start:    fyne.NewPos(0, 0),
			end:      fyne.NewPos(100, 100),
			expected: 14.142, // Distance to end point
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := perpendicularDistance(tt.point, tt.start, tt.end)
			assert.InDelta(t, tt.expected, result, 0.01, "Distance calculation mismatch")
		})
	}
}

// TestDistanceFunction tests Euclidean distance calculation
func TestDistanceFunction(t *testing.T) {
	tests := []struct {
		name     string
		p1       fyne.Position
		p2       fyne.Position
		expected float32
	}{
		{"Same point", fyne.NewPos(0, 0), fyne.NewPos(0, 0), 0},
		{"Horizontal", fyne.NewPos(0, 0), fyne.NewPos(10, 0), 10},
		{"Vertical", fyne.NewPos(0, 0), fyne.NewPos(0, 10), 10},
		{"Diagonal 3-4-5", fyne.NewPos(0, 0), fyne.NewPos(3, 4), 5},
		{"Negative coords", fyne.NewPos(-5, -5), fyne.NewPos(5, 5), 14.142},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := distance(tt.p1, tt.p2)
			assert.InDelta(t, tt.expected, result, 0.01)
		})
	}
}
