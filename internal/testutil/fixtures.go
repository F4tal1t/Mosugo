// Package testutil provides utility functions and sample data for testing the Mosugo application.
package testutil

import (
	"time"

	"mosugo/internal/storage"

	"fyne.io/fyne/v2"
)

// SampleWorkspaceEmpty returns an empty workspace state
func SampleWorkspaceEmpty() storage.WorkspaceState {
	return storage.WorkspaceState{
		Scale:   1.0,
		OffsetX: 0,
		OffsetY: 0,
		Cards:   []storage.MosuData{},
		Strokes: []storage.StrokeData{},
		Date:    "2024-01-01",
	}
}

// SampleWorkspaceSmall returns a workspace with a few cards and strokes
func SampleWorkspaceSmall() storage.WorkspaceState {
	return storage.WorkspaceState{
		Scale:   1.5,
		OffsetX: 100,
		OffsetY: 200,
		Cards: []storage.MosuData{
			{
				ID:        "card1",
				Content:   "- Task 1\n[] Checkbox",
				PosX:      100,
				PosY:      200,
				Width:     300,
				Height:    200,
				ColorIdx:  0,
				CreatedAt: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
			},
			{
				ID:        "card2",
				Content:   "Note",
				PosX:      500,
				PosY:      300,
				Width:     200,
				Height:    150,
				ColorIdx:  1,
				CreatedAt: time.Date(2024, 1, 1, 11, 0, 0, 0, time.UTC),
			},
		},
		Strokes: []storage.StrokeData{
			{P1X: 0, P1Y: 0, P2X: 100, P2Y: 100, ColorIdx: 0, Width: 2.0, StrokeID: 1},
			{P1X: 100, P1Y: 100, P2X: 200, P2Y: 150, ColorIdx: 0, Width: 2.0, StrokeID: 1},
			{P1X: 300, P1Y: 300, P2X: 400, P2Y: 400, ColorIdx: 1, Width: 3.0, StrokeID: 2},
		},
		Date: "2024-01-01",
	}
}

// SampleWorkspaceLarge returns a workspace with many cards and strokes for performance testing
func SampleWorkspaceLarge(cardCount, strokeCount int) storage.WorkspaceState {
	cards := make([]storage.MosuData, cardCount)
	for i := 0; i < cardCount; i++ {
		cards[i] = storage.MosuData{
			ID:        "card" + string(rune(i)),
			Content:   "Card content " + string(rune(i)),
			PosX:      float32(i * 50),
			PosY:      float32(i * 50),
			Width:     300,
			Height:    200,
			ColorIdx:  i % 5,
			CreatedAt: time.Now(),
		}
	}

	strokes := make([]storage.StrokeData, strokeCount)
	for i := 0; i < strokeCount; i++ {
		strokes[i] = storage.StrokeData{
			P1X:      float32(i * 10),
			P1Y:      float32(i * 10),
			P2X:      float32(i*10 + 50),
			P2Y:      float32(i*10 + 50),
			ColorIdx: i % 3,
			Width:    2.0,
			StrokeID: i/10 + 1, // Group every 10 segments
		}
	}

	return storage.WorkspaceState{
		Scale:   1.0,
		OffsetX: 0,
		OffsetY: 0,
		Cards:   cards,
		Strokes: strokes,
		Date:    "2024-01-01",
	}
}

// SampleStrokePoints returns sample stroke points for testing simplification
func SampleStrokePoints() struct {
	StraightLine []fyne.Position
	Zigzag       []fyne.Position
	Circle       []fyne.Position
	SinglePoint  []fyne.Position
	TwoPoints    []fyne.Position
} {
	// Straight line (should be reduced to 2 points)
	straightLine := make([]fyne.Position, 100)
	for i := 0; i < 100; i++ {
		straightLine[i] = fyne.Position{X: float32(i), Y: float32(i)}
	}

	// Zigzag pattern
	zigzag := make([]fyne.Position, 50)
	for i := 0; i < 50; i++ {
		y := float32(0)
		if i%2 == 0 {
			y = 10
		}
		zigzag[i] = fyne.Position{X: float32(i * 2), Y: y}
	}

	// Circle approximation
	circle := make([]fyne.Position, 100)
	for i := 0; i < 100; i++ {
		angle := float32(i) * 2 * 3.14159 / 100
		circle[i] = fyne.Position{
			X: 100 + 50*float32(cosApprox(angle)),
			Y: 100 + 50*float32(sinApprox(angle)),
		}
	}

	return struct {
		StraightLine []fyne.Position
		Zigzag       []fyne.Position
		Circle       []fyne.Position
		SinglePoint  []fyne.Position
		TwoPoints    []fyne.Position
	}{
		StraightLine: straightLine,
		Zigzag:       zigzag,
		Circle:       circle,
		SinglePoint:  []fyne.Position{{X: 10, Y: 20}},
		TwoPoints:    []fyne.Position{{X: 0, Y: 0}, {X: 100, Y: 100}},
	}
}

// Simple approximations for circle generation without importing math
func cosApprox(x float32) float32 {
	// Simple Taylor series approximation
	x2 := x * x
	return 1 - x2/2 + x2*x2/24
}

func sinApprox(x float32) float32 {
	x2 := x * x
	return x - x*x2/6 + x*x2*x2/120
}
