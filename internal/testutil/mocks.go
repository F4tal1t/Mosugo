package testutil

import (
	"github.com/F4tal1t/Mosugo/internal/cards"
	"github.com/F4tal1t/Mosugo/internal/storage"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"github.com/stretchr/testify/mock"
)

// MockCanvas is a mock implementation of the Canvas interface for testing
type MockCanvas struct {
	mock.Mock
}

func (m *MockCanvas) Refresh() {
	m.Called()
}

func (m *MockCanvas) ScreenToWorld(pos fyne.Position) fyne.Position {
	args := m.Called(pos)
	return args.Get(0).(fyne.Position)
}

func (m *MockCanvas) WorldToScreen(pos fyne.Position) fyne.Position {
	args := m.Called(pos)
	return args.Get(0).(fyne.Position)
}

func (m *MockCanvas) Snap(v float32) float32 {
	args := m.Called(v)
	return args.Get(0).(float32)
}

func (m *MockCanvas) SnapUp(v float32) float32 {
	args := m.Called(v)
	return args.Get(0).(float32)
}

func (m *MockCanvas) GetOffset() fyne.Position {
	args := m.Called()
	return args.Get(0).(fyne.Position)
}

func (m *MockCanvas) SetOffset(pos fyne.Position) {
	m.Called(pos)
}

func (m *MockCanvas) GetScale() float32 {
	args := m.Called()
	return args.Get(0).(float32)
}

func (m *MockCanvas) AddObject(o fyne.CanvasObject) {
	m.Called(o)
}

func (m *MockCanvas) RemoveObject(o fyne.CanvasObject) {
	m.Called(o)
}

func (m *MockCanvas) ContentContainer() *fyne.Container {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*fyne.Container)
}

func (m *MockCanvas) AddStroke(p1, p2 fyne.Position, strokeID int) {
	m.Called(p1, p2, strokeID)
}

func (m *MockCanvas) CollectCardData(card *cards.MosuWidget) storage.MosuData {
	args := m.Called(card)
	if args.Get(0) == nil {
		return storage.MosuData{}
	}
	return args.Get(0).(storage.MosuData)
}

func (m *MockCanvas) CollectStrokeDataByID(strokeID int) []storage.StrokeData {
	args := m.Called(strokeID)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]storage.StrokeData)
}

func (m *MockCanvas) CommitCardCreated(card *cards.MosuWidget) {
	m.Called(card)
}

func (m *MockCanvas) CommitCardDeleted(data storage.MosuData) {
	m.Called(data)
}

func (m *MockCanvas) CommitCardMoved(card *cards.MosuWidget, before fyne.Position) {
	m.Called(card, before)
}

func (m *MockCanvas) CommitStrokeCreated(segments []storage.StrokeData) {
	m.Called(segments)
}

func (m *MockCanvas) CommitStrokeDeleted(segments []storage.StrokeData) {
	m.Called(segments)
}

func (m *MockCanvas) MarkDirty() {
	m.Called()
}

func (m *MockCanvas) GhostRect() *canvas.Rectangle {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*canvas.Rectangle)
}

func (m *MockCanvas) GenerateStrokeID() int {
	args := m.Called()
	return args.Int(0)
}

func (m *MockCanvas) ValidateStrokeID(strokeID int) bool {
	args := m.Called(strokeID)
	return args.Bool(0)
}

func (m *MockCanvas) GetStrokeID(line *canvas.Line) (int, bool) {
	args := m.Called(line)
	return args.Int(0), args.Bool(1)
}

func (m *MockCanvas) IsGlowLine(line *canvas.Line) bool {
	args := m.Called(line)
	return args.Bool(0)
}

func (m *MockCanvas) SimplifyStroke(points []fyne.Position, epsilon float32) []fyne.Position {
	args := m.Called(points, epsilon)
	return args.Get(0).([]fyne.Position)
}

func (m *MockCanvas) GetStrokePoints(line *canvas.Line) (fyne.Position, fyne.Position, bool) {
	args := m.Called(line)
	return args.Get(0).(fyne.Position), args.Get(1).(fyne.Position), args.Bool(2)
}

func (m *MockCanvas) GetSelectedCard() *cards.MosuWidget {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*cards.MosuWidget)
}

func (m *MockCanvas) SetSelectedCard(c *cards.MosuWidget) {
	m.Called(c)
}
