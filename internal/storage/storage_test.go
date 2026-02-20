package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create unique test dates to avoid conflicts
func getTestDate(day int) time.Time {
	// Use year 2099 to avoid conflicts with real data
	return time.Date(2099, 1, day, 0, 0, 0, 0, time.UTC)
}

// TestSaveAndLoadWorkspaceEmpty tests empty workspace roundtrip
func TestSaveAndLoadWorkspaceEmpty(t *testing.T) {
	testDate := getTestDate(1)
	defer DeleteWorkspace(testDate)
	DeleteWorkspace(testDate)

	empty := WorkspaceState{
		Scale:   1.0,
		OffsetX: 0,
		OffsetY: 0,
		Cards:   []MosuData{},
		Strokes: []StrokeData{},
	}

	err := SaveWorkspace(testDate, empty)
	require.NoError(t, err, "Should save empty workspace")

	loaded, err := LoadWorkspace(testDate)
	require.NoError(t, err, "Should load empty workspace")

	assert.Equal(t, empty.Scale, loaded.Scale)
	assert.Equal(t, empty.OffsetX, loaded.OffsetX)
	assert.Equal(t, empty.OffsetY, loaded.OffsetY)
	assert.Len(t, loaded.Cards, 0)
	assert.Len(t, loaded.Strokes, 0)
}

// TestSaveAndLoadWorkspaceSmall tests small workspace with cards and strokes
func TestSaveAndLoadWorkspaceSmall(t *testing.T) {
	testDate := getTestDate(2)
	defer DeleteWorkspace(testDate)
	DeleteWorkspace(testDate)

	small := WorkspaceState{
		Scale:   1.5,
		OffsetX: 100,
		OffsetY: 200,
		Cards: []MosuData{
			{
				ID:        "card1",
				Content:   "Test Card 1",
				PosX:      100,
				PosY:      200,
				Width:     300,
				Height:    200,
				ColorIdx:  0,
				CreatedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			},
		},
		Strokes: []StrokeData{
			{P1X: 0, P1Y: 0, P2X: 100, P2Y: 100, ColorIdx: 0, Width: 2.0, StrokeID: 1},
		},
	}

	err := SaveWorkspace(testDate, small)
	require.NoError(t, err)

	loaded, err := LoadWorkspace(testDate)
	require.NoError(t, err)

	assert.Equal(t, small.Scale, loaded.Scale)
	assert.Len(t, loaded.Cards, 1)
	assert.Len(t, loaded.Strokes, 1)
	assert.Equal(t, small.Cards[0].ID, loaded.Cards[0].ID)
	assert.Equal(t, small.Cards[0].Content, loaded.Cards[0].Content)
}

// TestSaveWorkspaceWithUnicode tests unicode content in cards
func TestSaveWorkspaceWithUnicode(t *testing.T) {
	testDate := getTestDate(4)
	defer DeleteWorkspace(testDate)
	DeleteWorkspace(testDate)

	ws := WorkspaceState{
		Scale: 1.0,
		Cards: []MosuData{
			{
				ID:        "unicode-card",
				Content:   "Hello 世界 🌍 émojis",
				PosX:      100,
				PosY:      200,
				Width:     300,
				Height:    200,
				ColorIdx:  0,
				CreatedAt: time.Now(),
			},
		},
	}

	err := SaveWorkspace(testDate, ws)
	require.NoError(t, err)

	loaded, err := LoadWorkspace(testDate)
	require.NoError(t, err)

	assert.Equal(t, ws.Cards[0].Content, loaded.Cards[0].Content)
}

// TestLoadWorkspaceNotExists tests loading non-existent workspace returns empty
func TestLoadWorkspaceNotExists(t *testing.T) {
	testDate := getTestDate(99)
	DeleteWorkspace(testDate)

	loaded, err := LoadWorkspace(testDate)
	require.NoError(t, err, "Should return empty workspace, not error")
	assert.Len(t, loaded.Cards, 0)
	assert.Len(t, loaded.Strokes, 0)
	assert.Equal(t, float32(1.0), loaded.Scale)
}

// TestLoadWorkspaceCorruptJSON tests loading corrupted JSON file
func TestLoadWorkspaceCorruptJSON(t *testing.T) {
	testDate := getTestDate(5)
	defer DeleteWorkspace(testDate)

	storagePath, err := GetStoragePath()
	require.NoError(t, err)

	filePath := filepath.Join(storagePath, testDate.Format("2006-01-02")+".mosugo")
	err = os.WriteFile(filePath, []byte("{ invalid json }"), 0644)
	require.NoError(t, err)

	_, err = LoadWorkspace(testDate)
	assert.Error(t, err, "Should error on corrupted JSON")
}

// TestLoadWorkspaceLargeCoordinates tests extreme float32 values
func TestLoadWorkspaceLargeCoordinates(t *testing.T) {
	testDate := getTestDate(6)
	defer DeleteWorkspace(testDate)
	DeleteWorkspace(testDate)

	ws := WorkspaceState{
		Scale:   1.0,
		OffsetX: 1e10,
		OffsetY: -1e10,
		Cards: []MosuData{
			{
				ID:        "extreme-card",
				Content:   "Far away",
				PosX:      1e8,
				PosY:      -1e8,
				Width:     300,
				Height:    200,
				ColorIdx:  0,
				CreatedAt: time.Now(),
			},
		},
	}

	err := SaveWorkspace(testDate, ws)
	require.NoError(t, err)

	loaded, err := LoadWorkspace(testDate)
	require.NoError(t, err)

	assert.Equal(t, ws.OffsetX, loaded.OffsetX)
	assert.Equal(t, ws.OffsetY, loaded.OffsetY)
	assert.Equal(t, ws.Cards[0].PosX, loaded.Cards[0].PosX)
}

// TestListSavedDates tests listing saved workspaces
func TestListSavedDates(t *testing.T) {
	dates := []time.Time{getTestDate(10), getTestDate(11), getTestDate(12)}
	for _, date := range dates {
		defer DeleteWorkspace(date)
		DeleteWorkspace(date)

		ws := WorkspaceState{Scale: 1.0}
		err := SaveWorkspace(date, ws)
		require.NoError(t, err)
	}

	listed, err := ListSavedDates()
	require.NoError(t, err)

	// Check that our test dates are in the list
	foundCount := 0
	for _, listedDate := range listed {
		for _, testDate := range dates {
			if listedDate.Format("2006-01-02") == testDate.Format("2006-01-02") {
				foundCount++
			}
		}
	}
	assert.Equal(t, 3, foundCount, "Should find all 3 test dates")
}

// TestWorkspaceExists tests file existence checking
func TestWorkspaceExists(t *testing.T) {
	testDate := getTestDate(15)
	defer DeleteWorkspace(testDate)
	DeleteWorkspace(testDate)

	exists := WorkspaceExists(testDate)
	assert.False(t, exists)

	ws := WorkspaceState{Scale: 1.0}
	err := SaveWorkspace(testDate, ws)
	require.NoError(t, err)

	exists = WorkspaceExists(testDate)
	assert.True(t, exists)
}

// TestDeleteWorkspace tests workspace deletion
func TestDeleteWorkspace(t *testing.T) {
	testDate := getTestDate(16)

	ws := WorkspaceState{Scale: 1.0}
	err := SaveWorkspace(testDate, ws)
	require.NoError(t, err)

	assert.True(t, WorkspaceExists(testDate))

	err = DeleteWorkspace(testDate)
	require.NoError(t, err)

	assert.False(t, WorkspaceExists(testDate))
}

// TestDeleteWorkspaceNotExists tests deleting non-existent workspace
func TestDeleteWorkspaceNotExists(t *testing.T) {
	testDate := getTestDate(98)
	DeleteWorkspace(testDate)

	err := DeleteWorkspace(testDate)
	assert.NoError(t, err, "Deleting non-existent workspace should not error")
}

// TestGetStoragePath tests storage path retrieval
func TestGetStoragePath(t *testing.T) {
	path, err := GetStoragePath()
	require.NoError(t, err)
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "Mosugo")

	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

// TestConversionFunctions tests position and size conversion helpers
func TestConversionFunctions(t *testing.T) {
	// Test position conversion
	pos := ConvertStorageToPosition(10.5, 20.5)
	assert.Equal(t, float32(10.5), pos.X)
	assert.Equal(t, float32(20.5), pos.Y)

	x, y := ConvertPositionToStorage(pos)
	assert.Equal(t, float32(10.5), x)
	assert.Equal(t, float32(20.5), y)

	// Test size conversion
	size := ConvertStorageToSize(300, 200)
	assert.Equal(t, float32(300), size.Width)
	assert.Equal(t, float32(200), size.Height)

	w, h := ConvertSizeToStorage(size)
	assert.Equal(t, float32(300), w)
	assert.Equal(t, float32(200), h)
}
