package canvas

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"github.com/F4tal1t/Mosugo/internal/cards"
	"github.com/F4tal1t/Mosugo/internal/storage"
	"github.com/F4tal1t/Mosugo/internal/theme"
)

type historyCommand interface {
	Apply(c *MosugoCanvas)
	Undo(c *MosugoCanvas)
}

type cardCreateCommand struct {
	data storage.MosuData
}

func (cmd cardCreateCommand) Apply(c *MosugoCanvas) {
	c.addCardFromData(cmd.data)
	c.refreshIfReady()
}

func (cmd cardCreateCommand) Undo(c *MosugoCanvas) {
	c.removeCardByID(cmd.data.ID)
	c.refreshIfReady()
}

type cardDeleteCommand struct {
	data storage.MosuData
}

func (cmd cardDeleteCommand) Apply(c *MosugoCanvas) {
	c.removeCardByID(cmd.data.ID)
	c.refreshIfReady()
}

func (cmd cardDeleteCommand) Undo(c *MosugoCanvas) {
	c.addCardFromData(cmd.data)
	c.refreshIfReady()
}

type cardMoveCommand struct {
	cardID string
	before fyne.Position
	after  fyne.Position
}

func (cmd cardMoveCommand) Apply(c *MosugoCanvas) {
	if card := c.findCardByID(cmd.cardID); card != nil {
		card.WorldPos = cmd.after
		c.refreshIfReady()
	}
}

func (cmd cardMoveCommand) Undo(c *MosugoCanvas) {
	if card := c.findCardByID(cmd.cardID); card != nil {
		card.WorldPos = cmd.before
		c.refreshIfReady()
	}
}

type cardTextCommand struct {
	cardID string
	before string
	after  string
}

func (cmd cardTextCommand) Apply(c *MosugoCanvas) {
	if card := c.findCardByID(cmd.cardID); card != nil {
		card.SetText(cmd.after)
		card.RefreshContent()
		c.refreshIfReady()
	}
}

func (cmd cardTextCommand) Undo(c *MosugoCanvas) {
	if card := c.findCardByID(cmd.cardID); card != nil {
		card.SetText(cmd.before)
		card.RefreshContent()
		c.refreshIfReady()
	}
}

type strokeCreateCommand struct {
	segments []storage.StrokeData
}

func (cmd strokeCreateCommand) Apply(c *MosugoCanvas) {
	c.addStrokeSegments(cmd.segments)
	c.refreshIfReady()
}

func (cmd strokeCreateCommand) Undo(c *MosugoCanvas) {
	for _, segment := range cmd.segments {
		c.removeStrokeByID(segment.StrokeID)
	}
	c.refreshIfReady()
}

type strokeDeleteCommand struct {
	segments []storage.StrokeData
}

func (cmd strokeDeleteCommand) Apply(c *MosugoCanvas) {
	for _, segment := range cmd.segments {
		c.removeStrokeByID(segment.StrokeID)
	}
	c.refreshIfReady()
}

func (cmd strokeDeleteCommand) Undo(c *MosugoCanvas) {
	c.addStrokeSegments(cmd.segments)
	c.refreshIfReady()
}

func cloneStrokeSegments(segments []storage.StrokeData) []storage.StrokeData {
	if len(segments) == 0 {
		return nil
	}
	cloned := make([]storage.StrokeData, len(segments))
	copy(cloned, segments)
	return cloned
}

func (c *MosugoCanvas) resetHistory() {
	c.undoStack = nil
	c.redoStack = nil
}

func (c *MosugoCanvas) commitCommand(cmd historyCommand) {
	if c.suppressHistory {
		return
	}
	c.undoStack = append(c.undoStack, cmd)
	c.redoStack = nil
	c.notifyDirty()
}

func (c *MosugoCanvas) refreshIfReady() {
	if c.uiReady {
		c.Refresh()
	}
}

// CommitCardCreated records a newly created card as a reversible command.
func (c *MosugoCanvas) CommitCardCreated(card *cards.MosuWidget) {
	if card == nil {
		return
	}
	c.commitCommand(cardCreateCommand{data: c.CollectCardData(card)})
}

// CommitCardDeleted records a removed card as a reversible command.
func (c *MosugoCanvas) CommitCardDeleted(data storage.MosuData) {
	c.commitCommand(cardDeleteCommand{data: data})
}

// CommitCardMoved records a completed card move as a reversible command.
func (c *MosugoCanvas) CommitCardMoved(card *cards.MosuWidget, before fyne.Position) {
	if card == nil {
		return
	}
	if before == card.WorldPos {
		return
	}
	c.commitCommand(cardMoveCommand{cardID: card.ID, before: before, after: card.WorldPos})
}

// CommitCardTextChanged records a completed card text edit as a reversible command.
func (c *MosugoCanvas) CommitCardTextChanged(card *cards.MosuWidget, before, after string) {
	if card == nil || before == after {
		return
	}
	c.commitCommand(cardTextCommand{cardID: card.ID, before: before, after: after})
}

// CommitStrokeCreated records a completed stroke as a reversible command.
func (c *MosugoCanvas) CommitStrokeCreated(segments []storage.StrokeData) {
	if len(segments) == 0 {
		return
	}
	c.commitCommand(strokeCreateCommand{segments: cloneStrokeSegments(segments)})
}

// CommitStrokeDeleted records a removed stroke as a reversible command.
func (c *MosugoCanvas) CommitStrokeDeleted(segments []storage.StrokeData) {
	if len(segments) == 0 {
		return
	}
	c.commitCommand(strokeDeleteCommand{segments: cloneStrokeSegments(segments)})
}

// Undo reverts the latest committed command.
func (c *MosugoCanvas) Undo() bool {
	if len(c.undoStack) == 0 {
		return false
	}

	last := c.undoStack[len(c.undoStack)-1]
	c.undoStack = c.undoStack[:len(c.undoStack)-1]

	c.suppressHistory = true
	last.Undo(c)
	c.suppressHistory = false

	c.redoStack = append(c.redoStack, last)
	c.notifyDirty()
	return true
}

// Redo reapplies the latest undone command.
func (c *MosugoCanvas) Redo() bool {
	if len(c.redoStack) == 0 {
		return false
	}

	last := c.redoStack[len(c.redoStack)-1]
	c.redoStack = c.redoStack[:len(c.redoStack)-1]

	c.suppressHistory = true
	last.Apply(c)
	c.suppressHistory = false

	c.undoStack = append(c.undoStack, last)
	c.notifyDirty()
	return true
}

func (c *MosugoCanvas) findCardByID(id string) *cards.MosuWidget {
	for _, obj := range c.Content.Objects {
		card, ok := obj.(*cards.MosuWidget)
		if ok && card.ID == id {
			return card
		}
	}
	return nil
}

// CollectCardData captures the serializable state of a card.
func (c *MosugoCanvas) CollectCardData(card *cards.MosuWidget) storage.MosuData {
	if card == nil {
		return storage.MosuData{}
	}
	return storage.MosuData{
		ID:        card.ID,
		Content:   card.GetText(),
		PosX:      card.WorldPos.X,
		PosY:      card.WorldPos.Y,
		Width:     card.WorldSize.Width,
		Height:    card.WorldSize.Height,
		ColorIdx:  card.ColorIndex,
		CreatedAt: card.CreatedAt,
	}
}

func (c *MosugoCanvas) addCardFromData(data storage.MosuData) *cards.MosuWidget {
	card := cards.NewMosuWidget(data.ID, theme.CardBg, data.ColorIdx)
	card.WorldPos = fyne.NewPos(data.PosX, data.PosY)
	card.WorldSize = fyne.NewSize(data.Width, data.Height)
	card.CreatedAt = data.CreatedAt
	card.SetText(data.Content)
	card.RefreshContent()
	c.wireCardCallbacks(card)
	c.Content.Add(card)
	return card
}

func (c *MosugoCanvas) removeCardByID(id string) {
	for _, obj := range c.Content.Objects {
		card, ok := obj.(*cards.MosuWidget)
		if ok && card.ID == id {
			c.Content.Remove(obj)
			if c.selectedCard == card {
				c.selectedCard = nil
			}
			return
		}
	}
}

func (c *MosugoCanvas) wireCardCallbacks(card *cards.MosuWidget) {
	if card == nil {
		return
	}
	card.SetOnDirty(c.MarkDirty)
	card.SetOnTextCommitted(func(before, after string) {
		c.CommitCardTextChanged(card, before, after)
	})
}

func (c *MosugoCanvas) addStrokeSegments(segments []storage.StrokeData) {
	for _, segment := range segments {
		c.addStrokeSegment(segment)
	}
}

func (c *MosugoCanvas) addStrokeSegment(segment storage.StrokeData) {
	width := segment.Width
	if width <= 0 {
		width = c.StrokeWidth
	}

	previousWidth := c.StrokeWidth
	c.StrokeWidth = width
	c.AddStroke(fyne.NewPos(segment.P1X, segment.P1Y), fyne.NewPos(segment.P2X, segment.P2Y), segment.StrokeID)
	c.StrokeWidth = previousWidth
}

// CollectStrokeDataByID gathers the visible segments for a stroke ID.
func (c *MosugoCanvas) CollectStrokeDataByID(strokeID int) []storage.StrokeData {
	segments := []storage.StrokeData{}
	for line, coords := range c.strokesMap {
		if c.glowLines[line] {
			continue
		}
		id, ok := c.GetStrokeID(line)
		if !ok || id != strokeID {
			continue
		}
		segments = append(segments, storage.StrokeData{
			P1X:      coords.P1.X,
			P1Y:      coords.P1.Y,
			P2X:      coords.P2.X,
			P2Y:      coords.P2.Y,
			ColorIdx: 0,
			Width:    c.StrokeWidth,
			StrokeID: strokeID,
		})
	}
	return segments
}

func (c *MosugoCanvas) removeStrokeByID(strokeID int) {
	objectsToRemove := []fyne.CanvasObject{}
	for _, obj := range c.Content.Objects {
		line, ok := obj.(*canvas.Line)
		if !ok {
			continue
		}
		id, exists := c.GetStrokeID(line)
		if exists && id == strokeID {
			objectsToRemove = append(objectsToRemove, line)
		}
	}

	for _, obj := range objectsToRemove {
		c.RemoveObject(obj)
	}
}
