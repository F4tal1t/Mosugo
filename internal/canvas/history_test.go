package canvas

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"mosugo/internal/cards"
	"mosugo/internal/theme"
)

func countStrokeLines(c *MosugoCanvas, strokeID int) int {
	count := 0
	for _, obj := range c.Content.Objects {
		line, ok := obj.(*canvas.Line)
		if !ok {
			continue
		}
		id, exists := c.GetStrokeID(line)
		if exists && id == strokeID {
			count++
		}
	}
	return count
}

func TestUndoRedoCommands(t *testing.T) {
	c := NewMosugoCanvas()

	strokeID := c.GenerateStrokeID()
	p1 := fyne.NewPos(10, 20)
	p2 := fyne.NewPos(80, 90)
	c.AddStroke(p1, p2, strokeID)
	c.CommitStrokeCreated(c.CollectStrokeDataByID(strokeID))

	require.True(t, c.Undo(), "undo should restore the previous snapshot")
	assert.Equal(t, 0, countStrokeLines(c, strokeID))

	require.True(t, c.Redo(), "redo should restore the reverted snapshot")
	assert.Greater(t, countStrokeLines(c, strokeID), 0)
}

func TestFocusedCardShortcutForwarding(t *testing.T) {
	card := cards.NewMosuWidget("card-1", theme.CardBg, 0)
	undoCount := 0

	card.SetOnShortcut(func(shortcut fyne.Shortcut) {
		if custom, ok := shortcut.(*desktop.CustomShortcut); ok && custom.KeyName == fyne.KeyZ {
			undoCount++
		}
	})

	card.TypedShortcut(&desktop.CustomShortcut{KeyName: fyne.KeyZ, Modifier: fyne.KeyModifierControl})

	assert.Equal(t, 1, undoCount)
}

func TestCardTextEditsCommitImmediately(t *testing.T) {
	c := NewMosugoCanvas()
	card := cards.NewMosuWidget("card-edit", theme.CardBg, 0)
	c.AddObject(card)

	card.TypedRune('a')
	card.TypedKey(&fyne.KeyEvent{Name: fyne.KeyBackspace})
	card.TypedRune('b')

	require.True(t, c.Undo())
	assert.Equal(t, "", card.GetText())

	require.True(t, c.Redo())
	assert.Equal(t, "b", card.GetText())
}
