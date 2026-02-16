package tools

type ToolType int

const (
	ToolPan ToolType = iota
	ToolSelect
	ToolCard
	ToolDraw
	ToolErase
)

func (t ToolType) String() string {
	switch t {
	case ToolPan:
		return "Pan Mode"
	case ToolSelect:
		return "Select Mode"
	case ToolCard:
		return "Card Mode"
	case ToolDraw:
		return "Draw Mode"
	case ToolErase:
		return "Erase Mode"
	default:
		return "Unknown"
	}
}
