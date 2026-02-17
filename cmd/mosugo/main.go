package main

import (
	"fmt"
	"image/color"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	mosuCanvas "mosugo/internal/canvas"
	"mosugo/internal/theme"
	"mosugo/internal/tools"
	"mosugo/internal/ui"
)

var (
	BorderColor = color.RGBA{240, 240, 245, 255}
)

func createToolButton(iconPath string, tool tools.ToolType, mosugoCanvas *mosuCanvas.MosugoCanvas) *widget.Button {
	var icon fyne.Resource
	if res, err := fyne.LoadResourceFromPath(iconPath); err == nil {
		icon = res
	} else {
		log.Println("Could not load icon:", iconPath, err)
	}

	btn := widget.NewButtonWithIcon("", icon, func() {
		mosugoCanvas.CurrentTool = tool
		mosugoCanvas.Refresh()
		fmt.Println("Selected Tool:", tool)
	})
	btn.Importance = widget.LowImportance
	return btn
}

func main() {
	a := app.NewWithID("com.mosugo")
	a.Settings().SetTheme(theme.NewMosugoTheme())

	w := a.NewWindow("Mosugo")
	w.Resize(fyne.NewSize(1200, 800))
	w.SetPadded(false)

	mosugoCanvas := mosuCanvas.NewMosugoCanvas()

	cardBtn := createToolButton("assets/card.svg", tools.ToolCard, mosugoCanvas)
	drawBtn := createToolButton("assets/draw.svg", tools.ToolDraw, mosugoCanvas)
	eraseBtn := createToolButton("assets/eraser.svg", tools.ToolErase, mosugoCanvas)

	// Toolbar Buttons
	toolbarButtons := container.NewVBox(
		cardBtn,
		drawBtn,
		eraseBtn,
	)

	// Create the Metaball Border Overlay
	metaBorder := ui.NewMetaballBorder(BorderColor)

	// 1. Canvas layer
	canvasLayer := mosugoCanvas

	// 2. Border Overlay (Frame + Tab Background)
	borderLayer := metaBorder

	// 3. Toolbar Buttons (Left Aligned)
	// We use padding to push it into the tab area properly

	leftPadding := canvas.NewRectangle(color.Transparent)
	leftPadding.SetMinSize(fyne.NewSize(12, 0))

	toolbarAligned := container.NewHBox(
		leftPadding,
		toolbarButtons,
		layout.NewSpacer(),
	)

	toolbarLayer := container.NewVBox(
		layout.NewSpacer(),
		toolbarAligned,
		layout.NewSpacer(),
	)

	finalLayout := container.NewStack(
		canvasLayer,
		borderLayer,
		toolbarLayer,
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

	w.SetContent(finalLayout)
	w.ShowAndRun()
}
