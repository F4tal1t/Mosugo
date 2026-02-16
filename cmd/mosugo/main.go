package main

import (
	"fmt"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"

	mosuCanvas "mosugo/internal/canvas"
	"mosugo/internal/theme"
	"mosugo/internal/tools"
)

func createIconButton(label string, icon string, tool tools.ToolType, mosugoCanvas *mosuCanvas.MosugoCanvas) *widget.Button {
	btn := widget.NewButton(icon, func() {
		mosugoCanvas.CurrentTool = tool
		mosugoCanvas.Refresh()
	})
	btn.Importance = widget.LowImportance
	return btn
}

func main() {
	a := app.NewWithID("com.mosugo")
	a.Settings().SetTheme(theme.NewMosugoTheme())

	w := a.NewWindow("Mosugo")
	w.Resize(fyne.NewSize(1000, 800))
	w.SetPadded(false)

	mosugoCanvas := mosuCanvas.NewMosugoCanvas()

	cardBtn := createIconButton("Card", "📋", tools.ToolCard, mosugoCanvas)
	drawBtn := createIconButton("Draw", "✏️", tools.ToolDraw, mosugoCanvas)
	eraseBtn := createIconButton("Erase", "🗑️", tools.ToolErase, mosugoCanvas)

	strokeWidthLabel := widget.NewLabel("Stroke")
	strokeWidthSlider := widget.NewSlider(1, 10)
	strokeWidthSlider.Value = 2
	strokeWidthSlider.Step = 0.5
	strokeWidthSlider.OnChanged = func(value float64) {
		mosugoCanvas.StrokeWidth = float32(value)
	}

	toolbarBg := canvas.NewRectangle(color.RGBA{245, 245, 248, 250})
	toolbarBg.CornerRadius = 12
	toolbarBg.StrokeColor = color.RGBA{220, 220, 230, 200}
	toolbarBg.StrokeWidth = 1

	toolbarButtons := container.NewVBox(
		cardBtn,
		drawBtn,
		eraseBtn,
		widget.NewSeparator(),
		strokeWidthLabel,
		strokeWidthSlider,
	)

	toolbarCard := container.NewStack(
		toolbarBg,
		container.NewPadded(toolbarButtons),
	)

	uiLayer := container.NewBorder(
		nil,
		nil,
		nil,
		container.NewPadded(
			container.NewVBox(
				toolbarCard,
			),
		),
	)

	if deskCanvas, ok := w.Canvas().(desktop.Canvas); ok {
		deskCanvas.SetOnKeyDown(func(key *fyne.KeyEvent) {
			switch key.Name {
			case fyne.Key1, "KP1":
				mosugoCanvas.CurrentTool = tools.ToolCard
				fmt.Println("Tool: Card Mode")
				mosugoCanvas.Refresh()
			case fyne.Key2, "KP2":
				mosugoCanvas.CurrentTool = tools.ToolDraw
				fmt.Println("Tool: Draw Mode")
				mosugoCanvas.Refresh()
			case fyne.Key3, "KP3":
				mosugoCanvas.CurrentTool = tools.ToolErase
				fmt.Println("Tool: Erase Mode")
				mosugoCanvas.Refresh()

			case fyne.KeyEscape:
				mosugoCanvas.CurrentTool = tools.ToolCard
				mosugoCanvas.Refresh()
			}
		})
	}

	content := container.NewStack(
		mosugoCanvas,
		uiLayer,
	)
	w.SetContent(content)

	w.ShowAndRun()
}
