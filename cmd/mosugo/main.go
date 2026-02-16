package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"mosugo/internal/canvas"
	"mosugo/internal/theme"
	"mosugo/internal/tools"
)

func main() {
	a := app.NewWithID("com.mosugo")
	// Set custom theme to allow later font customization
	a.Settings().SetTheme(theme.NewMosugoTheme())

	w := a.NewWindow("Mosugo")

	// Custom Icon (Placeholder logic)
	// w.SetIcon(resourceIconPng)

	w.Resize(fyne.NewSize(1000, 800))
	w.SetPadded(false)

	// 1. Initialize Canvas
	mosugoCanvas := canvas.NewMosugoCanvas()

	// 2. Create Toolbar Overlay
	toolbar := container.NewHBox(
		widget.NewButton("Pan (0)", func() {
			mosugoCanvas.CurrentTool = tools.ToolPan
			mosugoCanvas.Refresh()
		}),
		widget.NewButton("Select (1)", func() {
			mosugoCanvas.CurrentTool = tools.ToolSelect
			mosugoCanvas.Refresh()
		}),
		widget.NewButton("Card (2)", func() {
			mosugoCanvas.CurrentTool = tools.ToolCard
			mosugoCanvas.Refresh()
		}),
		widget.NewButton("Draw (3)", func() {
			mosugoCanvas.CurrentTool = tools.ToolDraw
			mosugoCanvas.Refresh()
		}),
	)

	// Wrap Toolbar in a Top-Left aligned container
	uiLayer := container.NewVBox(
		container.NewHBox(toolbar),
		layout.NewSpacer(),
	)

	// 3. Setup Input Handlers (Global Key Listener)
	if deskCanvas, ok := w.Canvas().(desktop.Canvas); ok {
		deskCanvas.SetOnKeyDown(func(key *fyne.KeyEvent) {
			switch key.Name {
			// Numpad Tool Switching (Map Standard Numbers too)
			case fyne.Key0, "KP0":
				mosugoCanvas.CurrentTool = tools.ToolPan
				fmt.Println("Tool: Pan Mode")
				mosugoCanvas.Refresh()
			case fyne.Key1, "KP1":
				mosugoCanvas.CurrentTool = tools.ToolSelect // Changed from Card -> Select to match sequence
				fmt.Println("Tool: Select Mode")
				mosugoCanvas.Refresh()
			case fyne.Key2, "KP2":
				mosugoCanvas.CurrentTool = tools.ToolCard
				fmt.Println("Tool: Card Mode")
				mosugoCanvas.Refresh()
			case fyne.Key3, "KP3":
				mosugoCanvas.CurrentTool = tools.ToolDraw
				fmt.Println("Tool: Draw Mode")
				mosugoCanvas.Refresh()
			case fyne.Key4, "KP4":
				mosugoCanvas.CurrentTool = tools.ToolErase
				fmt.Println("Tool: Erase Mode")
				mosugoCanvas.Refresh()

			// Zoom Controls (Testing)
			case fyne.KeyPlus, fyne.KeyEqual: // +
				mosugoCanvas.ApplyZoom(0.1)
				mosugoCanvas.Refresh()
				fmt.Printf("Zoom: %.2f\n", mosugoCanvas.Zoom)
			case fyne.KeyMinus: // -
				mosugoCanvas.ApplyZoom(-0.1)
				mosugoCanvas.Refresh()
				fmt.Printf("Zoom: %.2f\n", mosugoCanvas.Zoom)

			// Escape to Reset
			case fyne.KeyEscape:
				mosugoCanvas.CurrentTool = tools.ToolPan
				mosugoCanvas.Refresh()
			}
		})
	}
	// 4. Layout
	// Canvas fills the window, UI Layer sits "above" it
	content := container.NewStack(
		mosugoCanvas,
		uiLayer, // Use the predefined UI layer
	)
	w.SetContent(content)

	w.ShowAndRun()
}
