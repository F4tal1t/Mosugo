package canvas

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAddStrokeCreatesGlowAndRegularPair tests that AddStroke creates both lines
func TestAddStrokeCreatesGlowAndRegularPair(t *testing.T) {
	c := NewMosugoCanvas()

	initialObjects := len(c.Content.Objects)

	p1 := fyne.NewPos(10, 20)
	p2 := fyne.NewPos(100, 200)
	strokeID := c.GenerateStrokeID()

	c.AddStroke(p1, p2, strokeID)

	// Should add 2 lines: glow and regular
	assert.Equal(t, initialObjects+2, len(c.Content.Objects), "Should add 2 lines (glow + regular)")

	// Find the two new lines
	var glowLine, regularLine *canvas.Line
	for i := initialObjects; i < len(c.Content.Objects); i++ {
		if line, ok := c.Content.Objects[i].(*canvas.Line); ok {
			if c.IsGlowLine(line) {
				glowLine = line
			} else {
				regularLine = line
			}
		}
	}

	require.NotNil(t, glowLine, "Should find glow line")
	require.NotNil(t, regularLine, "Should find regular line")

	// Verify glow line has wider stroke
	assert.Greater(t, glowLine.StrokeWidth, regularLine.StrokeWidth, "Glow line should be wider")
	assert.Equal(t, c.StrokeWidth*1.5, glowLine.StrokeWidth)
	assert.Equal(t, c.StrokeWidth, regularLine.StrokeWidth)

	// Both should have same stroke ID
	glowID, ok := c.GetStrokeID(glowLine)
	require.True(t, ok)
	assert.Equal(t, strokeID, glowID)

	regularID, ok := c.GetStrokeID(regularLine)
	require.True(t, ok)
	assert.Equal(t, strokeID, regularID)
}

// TestAddStrokeWithInvalidIDGeneratesNew tests defensive ID generation
func TestAddStrokeWithInvalidIDGeneratesNew(t *testing.T) {
	c := NewMosugoCanvas()

	p1 := fyne.NewPos(0, 0)
	p2 := fyne.NewPos(100, 100)

	// Try to add stroke with invalid ID
	c.AddStroke(p1, p2, 0)

	// Should have generated a valid ID instead
	foundValidID := false
	for line := range c.strokeIDMap {
		id, ok := c.GetStrokeID(line)
		if ok && c.ValidateStrokeID(id) {
			foundValidID = true
			break
		}
	}

	assert.True(t, foundValidID, "Should have generated valid ID for invalid input")
}

// TestRemoveObjectCleansUpMaps tests that RemoveObject cleans stroke tracking
func TestRemoveObjectCleansUpMaps(t *testing.T) {
	c := NewMosugoCanvas()

	line := canvas.NewLine(nil)
	p1 := fyne.NewPos(10, 20)
	p2 := fyne.NewPos(100, 200)
	strokeID := 1

	// Register and add line
	c.RegisterStroke(line, p1, p2, strokeID)
	c.glowLines[line] = true
	c.Content.Add(line)

	// Verify it's tracked
	_, ok := c.GetStrokeID(line)
	require.True(t, ok)
	assert.True(t, c.IsGlowLine(line))

	// Remove it
	c.RemoveObject(line)

	// Verify cleanup
	_, ok = c.GetStrokeID(line)
	assert.False(t, ok, "Stroke ID should be removed from map")

	_, _, ok = c.GetStrokePoints(line)
	assert.False(t, ok, "Stroke coords should be removed from map")

	assert.False(t, c.IsGlowLine(line), "Glow line marker should be removed")

	// Verify not in container
	found := false
	for _, obj := range c.Content.Objects {
		if obj == line {
			found = true
			break
		}
	}
	assert.False(t, found, "Line should be removed from container")
}

// TestClearCanvasCompleteReset tests comprehensive canvas reset
func TestClearCanvasCompleteReset(t *testing.T) {
	c := NewMosugoCanvas()

	// Add some strokes
	for i := 0; i < 5; i++ {
		p1 := fyne.NewPos(float32(i*10), float32(i*10))
		p2 := fyne.NewPos(float32(i*10+50), float32(i*10+50))
		strokeID := c.GenerateStrokeID()
		c.AddStroke(p1, p2, strokeID)
	}

	// Verify strokes were added
	require.Greater(t, len(c.Content.Objects), 1, "Should have objects")
	require.Greater(t, len(c.strokesMap), 0, "Should have stroke coords")
	require.Greater(t, len(c.strokeIDMap), 0, "Should have stroke IDs")
	require.Greater(t, len(c.glowLines), 0, "Should have glow lines")

	// Store values before clear
	offsetBefore := c.Offset
	scaleBefore := c.Scale

	// Clear canvas
	c.ClearCanvas()

	// Verify maps are cleared
	assert.Empty(t, c.strokesMap, "Stroke coords map should be empty")
	assert.Empty(t, c.strokeIDMap, "Stroke ID map should be empty")
	assert.Empty(t, c.glowLines, "Glow lines map should be empty")
	assert.Empty(t, c.strokes, "Strokes array should be empty")
	assert.Empty(t, c.currentStroke, "Current stroke should be empty")

	// Verify nextStrokeID is reset
	assert.Equal(t, 1, c.nextStrokeID, "Next stroke ID should reset to 1")

	// Verify selected card is cleared
	assert.Nil(t, c.selectedCard, "Selected card should be nil")

	// Verify offset and scale are unchanged
	assert.Equal(t, offsetBefore, c.Offset, "Offset should be preserved")
	assert.Equal(t, scaleBefore, c.Scale, "Scale should be preserved")

	// Only ghost rect should remain
	nonGhostObjects := 0
	for _, obj := range c.Content.Objects {
		if obj != c.ghostRect {
			nonGhostObjects++
		}
	}
	assert.Equal(t, 0, nonGhostObjects, "Only ghost rect should remain")
}

// TestClearCanvasPreservesGhostRect ensures ghost rect isn't removed
func TestClearCanvasPreservesGhostRect(t *testing.T) {
	c := NewMosugoCanvas()
	ghostRect := c.ghostRect

	// Add strokes
	c.AddStroke(fyne.NewPos(0, 0), fyne.NewPos(100, 100), c.GenerateStrokeID())

	// Clear
	c.ClearCanvas()

	// Ghost rect should still be there
	found := false
	for _, obj := range c.Content.Objects {
		if obj == ghostRect {
			found = true
			break
		}
	}
	assert.True(t, found, "Ghost rect should be preserved")
	assert.Equal(t, ghostRect, c.ghostRect, "Ghost rect reference should be unchanged")
}

// TestStrokeIDMigrationOnLoad tests invalid ID migration
func TestStrokeIDMigrationOnLoad(t *testing.T) {
	c := NewMosugoCanvas()

	// Simulate loading strokes with invalid IDs (old format)
	invalidStroke1 := struct {
		p1       fyne.Position
		p2       fyne.Position
		strokeID int
	}{
		p1:       fyne.NewPos(0, 0),
		p2:       fyne.NewPos(10, 10),
		strokeID: 0, // Invalid
	}

	invalidStroke2 := struct {
		p1       fyne.Position
		p2       fyne.Position
		strokeID int
	}{
		p1:       fyne.NewPos(20, 20),
		p2:       fyne.NewPos(30, 30),
		strokeID: -1, // Invalid
	}

	// AddStroke should auto-fix invalid IDs
	c.AddStroke(invalidStroke1.p1, invalidStroke1.p2, invalidStroke1.strokeID)
	c.AddStroke(invalidStroke2.p1, invalidStroke2.p2, invalidStroke2.strokeID)

	// All strokes should now have valid IDs
	validCount := 0
	for line := range c.strokeIDMap {
		id, ok := c.GetStrokeID(line)
		if ok && c.ValidateStrokeID(id) {
			validCount++
		}
	}

	// Should have 4 lines (2 glow + 2 regular), all with valid IDs
	assert.Equal(t, 4, validCount, "All strokes should have valid IDs")
}

// TestAddMultipleStrokesSameID tests grouping strokes with same ID
func TestAddMultipleStrokesSameID(t *testing.T) {
	c := NewMosugoCanvas()

	sharedStrokeID := c.GenerateStrokeID()

	// Add multiple segments with same ID (simulating one continuous stroke)
	strokes := []struct {
		p1 fyne.Position
		p2 fyne.Position
	}{
		{fyne.NewPos(0, 0), fyne.NewPos(10, 10)},
		{fyne.NewPos(10, 10), fyne.NewPos(20, 15)},
		{fyne.NewPos(20, 15), fyne.NewPos(30, 10)},
	}

	for _, s := range strokes {
		c.AddStroke(s.p1, s.p2, sharedStrokeID)
	}

	// Count lines with this stroke ID
	count := 0
	for line := range c.strokeIDMap {
		id, ok := c.GetStrokeID(line)
		if ok && id == sharedStrokeID {
			count++
		}
	}

	// Should have 6 lines: 3 segments × 2 (glow + regular)
	assert.Equal(t, 6, count, "Should have 6 lines for 3 segments (glow + regular each)")
}

// TestContentContainerGetter tests ContentContainer method
func TestContentContainerGetter(t *testing.T) {
	c := NewMosugoCanvas()

	container := c.ContentContainer()

	assert.NotNil(t, container)
	assert.Equal(t, c.Content, container)
}

// TestGhostRectGetter tests GhostRect method
func TestGhostRectGetter(t *testing.T) {
	c := NewMosugoCanvas()

	rect := c.GhostRect()

	assert.NotNil(t, rect)
	assert.Equal(t, c.ghostRect, rect)
	assert.False(t, rect.Visible(), "Ghost rect should be hidden initially")
}

// TestAddObjectToContent tests AddObject method
func TestAddObjectToContent(t *testing.T) {
	c := NewMosugoCanvas()

	initialCount := len(c.Content.Objects)

	rect := canvas.NewRectangle(nil)
	c.AddObject(rect)

	assert.Equal(t, initialCount+1, len(c.Content.Objects))
	assert.Contains(t, c.Content.Objects, rect)
}

// TestRemoveNonLineObject tests removing non-line objects
func TestRemoveNonLineObject(t *testing.T) {
	c := NewMosugoCanvas()

	rect := canvas.NewRectangle(nil)
	c.Content.Add(rect)

	c.RemoveObject(rect)

	assert.NotContains(t, c.Content.Objects, rect)
}
