package testutil

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"github.com/stretchr/testify/assert"
)

const (
	// PositionTolerance is the acceptable floating-point error for coordinate comparisons
	PositionTolerance = 0.001
)

// PositionEqual asserts that two fyne.Position values are equal within tolerance
func PositionEqual(t *testing.T, expected, actual fyne.Position, msgAndArgs ...interface{}) bool {
	t.Helper()
	xEqual := math.Abs(float64(expected.X-actual.X)) < PositionTolerance
	yEqual := math.Abs(float64(expected.Y-actual.Y)) < PositionTolerance

	if !xEqual || !yEqual {
		return assert.Fail(t, "Positions not equal within tolerance",
			"Expected: (%f, %f)\nActual: (%f, %f)\nTolerance: %f",
			expected.X, expected.Y, actual.X, actual.Y, PositionTolerance)
	}
	return true
}

// Float32Equal asserts that two float32 values are equal within tolerance
func Float32Equal(t *testing.T, expected, actual float32, msgAndArgs ...interface{}) bool {
	t.Helper()
	if math.Abs(float64(expected-actual)) >= PositionTolerance {
		return assert.Fail(t, "Float32 values not equal within tolerance",
			"Expected: %f\nActual: %f\nTolerance: %f",
			expected, actual, PositionTolerance)
	}
	return true
}

// CreateTempStorage creates a temporary directory for storage tests
// Returns the path and a cleanup function
func CreateTempStorage(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// CreateTestWorkspaceFile creates a test workspace file in the given directory
func CreateTestWorkspaceFile(t *testing.T, dir, date, content string) string {
	t.Helper()
	filename := filepath.Join(dir, date+".mosugo")
	err := os.WriteFile(filename, []byte(content), 0644)
	assert.NoError(t, err)
	return filename
}

// SamplePositions returns common test positions for coordinate transform tests
func SamplePositions() []fyne.Position {
	return []fyne.Position{
		{X: 0, Y: 0},          // Origin
		{X: 100, Y: 100},      // Positive quadrant
		{X: -50, Y: -50},      // Negative quadrant
		{X: 1000, Y: 500},     // Large values
		{X: 0.5, Y: 0.5},      // Fractional
		{X: -100.5, Y: 200.3}, // Mixed signs with fractions
	}
}

// SampleScales returns common test scale values
func SampleScales() []float32 {
	return []float32{0.1, 0.5, 1.0, 2.0, 5.0, 10.0}
}
