// Package storage handles JSON-based persistence of workspace data.
// It implements a daily workspace model where each day's cards and strokes
// are saved to a separate JSON file, providing automatic versioning and
// easy navigation through time.
package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
)

// MosuData represents the serializable data for a Mosu card.
type MosuData struct {
	ID        string    `json:"id"`
	Content   string    `json:"content"`
	PosX      float32   `json:"pos_x"`
	PosY      float32   `json:"pos_y"`
	Width     float32   `json:"width"`
	Height    float32   `json:"height"`
	ColorIdx  int       `json:"color_index"`
	CreatedAt time.Time `json:"created_at"`
}

// StrokeData represents a drawing stroke with its world coordinates.
// Each stroke is a line segment defined by two points (P1, P2) with
// color, width, and a unique ID for grouping related segments.
type StrokeData struct {
	P1X      float32 `json:"p1_x"`
	P1Y      float32 `json:"p1_y"`
	P2X      float32 `json:"p2_x"`
	P2Y      float32 `json:"p2_y"`
	ColorIdx int     `json:"color_index"`
	Width    float32 `json:"width"`
	StrokeID int     `json:"stroke_id"`
}

// WorkspaceState represents the complete state of a workspace for a specific date.
type WorkspaceState struct {
	Scale   float32      `json:"scale"`
	OffsetX float32      `json:"offset_x"`
	OffsetY float32      `json:"offset_y"`
	Cards   []MosuData   `json:"cards"`
	Strokes []StrokeData `json:"strokes"`
	Date    string       `json:"date"` // YYYY-MM-DD format
}

// GetStoragePath returns the platform-specific storage directory for Mosugo workspaces.
// On Windows: %APPDATA%\Roaming\Mosugo\
// On Linux: ~/.config/Mosugo/
// The directory is created if it doesn't exist.
func GetStoragePath() (string, error) {
	// Get user config directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("failed to get config directory: %w", err)
	}

	// Append Mosugo subdirectory
	storagePath := filepath.Join(configDir, "Mosugo")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %w", err)
	}

	return storagePath, nil
}

// getWorkspaceFilePath returns the full path to a workspace file for a given date
func getWorkspaceFilePath(date time.Time) (string, error) {
	storagePath, err := GetStoragePath()
	if err != nil {
		return "", err
	}

	// Format: YYYY-MM-DD.mosugo
	filename := date.Format("2006-01-02") + ".mosugo"
	return filepath.Join(storagePath, filename), nil
}

// SaveWorkspace saves the current workspace state to a dated file
func SaveWorkspace(date time.Time, state WorkspaceState) error {
	// Ensure date field matches the file date
	state.Date = date.Format("2006-01-02")

	filePath, err := getWorkspaceFilePath(date)
	if err != nil {
		return err
	}

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workspace state: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write workspace file: %w", err)
	}

	return nil
}

// LoadWorkspace loads a workspace state from a dated file
func LoadWorkspace(date time.Time) (WorkspaceState, error) {
	filePath, err := getWorkspaceFilePath(date)
	if err != nil {
		return WorkspaceState{}, err
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// Return empty workspace with default values
		return WorkspaceState{
			Scale:   1.0,
			OffsetX: 0,
			OffsetY: 0,
			Cards:   []MosuData{},
			Strokes: []StrokeData{},
			Date:    date.Format("2006-01-02"),
		}, nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return WorkspaceState{}, fmt.Errorf("failed to read workspace file: %w", err)
	}

	// Unmarshal JSON
	var state WorkspaceState
	if err := json.Unmarshal(data, &state); err != nil {
		return WorkspaceState{}, fmt.Errorf("failed to unmarshal workspace state: %w", err)
	}

	return state, nil
}

func ListSavedDates() ([]time.Time, error) {
	storagePath, err := GetStoragePath()
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(storagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read storage directory: %w", err)
	}

	var dates []time.Time

	// Parse filenames for dates
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		name := file.Name()
		if !strings.HasSuffix(name, ".mosugo") {
			continue
		}

		// Extract date from filename (YYYY-MM-DD.mosugo)
		dateStr := strings.TrimSuffix(name, ".mosugo")
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			// Skip invalid filenames
			continue
		}

		dates = append(dates, date)
	}

	// Sort dates in descending order (newest first)
	sort.Slice(dates, func(i, j int) bool {
		return dates[i].After(dates[j])
	})

	return dates, nil
}

// WorkspaceExists checks if a workspace file exists for a given date
func WorkspaceExists(date time.Time) bool {
	filePath, err := getWorkspaceFilePath(date)
	if err != nil {
		return false
	}

	_, err = os.Stat(filePath)
	return !os.IsNotExist(err)
}

// DeleteWorkspace deletes the workspace file for a given date
func DeleteWorkspace(date time.Time) error {
	filePath, err := getWorkspaceFilePath(date)
	if err != nil {
		return err
	}

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete workspace file: %w", err)
	}

	return nil
}

// ConvertPositionToStorage converts a fyne.Position to separate X and Y floats
func ConvertPositionToStorage(pos fyne.Position) (float32, float32) {
	return pos.X, pos.Y
}

// ConvertStorageToPosition converts separate X and Y floats back to a fyne.Position
func ConvertStorageToPosition(x, y float32) fyne.Position {
	return fyne.NewPos(x, y)
}

// ConvertSizeToStorage converts a fyne.Size to separate Width and Height floats
func ConvertSizeToStorage(size fyne.Size) (float32, float32) {
	return size.Width, size.Height
}

// ConvertStorageToSize converts separate Width and Height floats back to a fyne.Size
func ConvertStorageToSize(width, height float32) fyne.Size {
	return fyne.NewSize(width, height)
}
