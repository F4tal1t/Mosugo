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
	BorderColor = color.RGBA{0, 31, 45, 255}
	toolButtons []*widget.Button
)

func createToolButton(iconPath string, tool tools.ToolType, mosugoCanvas *mosuCanvas.MosugoCanvas) *widget.Button {
	var icon fyne.Resource
	if res, err := fyne.LoadResourceFromPath(iconPath); err == nil {
		icon = res
	} else {
		log.Println("Could not load icon:", iconPath, err)
	}
	var btn *widget.Button
	btn = widget.NewButtonWithIcon("", icon, func() {
		mosugoCanvas.SetTool(tool)
		for _, b := range toolButtons {
			b.Importance = widget.LowImportance
			b.Refresh()
		}
		btn.Importance = widget.HighImportance
		btn.Refresh()
		fmt.Println("Selected Tool:", tool)
	})
	btn.Importance = widget.LowImportance
	toolButtons = append(toolButtons, btn)
	return btn
}

func main() {
	a := app.NewWithID("com.mosugo")
	a.Settings().SetTheme(theme.NewMosugoTheme())

	w := a.NewWindow("Mosugo")
	w.Resize(fyne.NewSize(600, 400))
	w.SetPadded(false)

	mosugoCanvas := mosuCanvas.NewMosugoCanvas()

	selectBtn := createToolButton("assets/select.svg", tools.ToolSelect, mosugoCanvas)
	cardBtn := createToolButton("assets/card.svg", tools.ToolCard, mosugoCanvas)
	drawBtn := createToolButton("assets/draw.svg", tools.ToolDraw, mosugoCanvas)
	eraseBtn := createToolButton("assets/eraser.svg", tools.ToolErase, mosugoCanvas)

	toolbarButtons := container.NewGridWrap(fyne.NewSize(35, 35),
		selectBtn,
		cardBtn,
		drawBtn,
		eraseBtn,
	)

	// creating Metaball Border Overlay
	metaBorder := ui.NewMetaballBorder(BorderColor)

	// canvas layer
	canvasLayer := mosugoCanvas

	// border Overlay (Frame + Tab Background)
	borderLayer := metaBorder

	leftPadding := canvas.NewRectangle(color.Transparent)
	leftPadding.SetMinSize(fyne.NewSize(5, 0))

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
			var targetBtn *widget.Button
			var targetTool tools.ToolType

			switch key.Name {
			case fyne.Key1, "KP1":
				mosugoCanvas.SetTool(tools.ToolCard)
				fmt.Println("Tool: Card Mode")
			case fyne.Key2, "KP2":
				mosugoCanvas.SetTool(tools.ToolDraw)
				fmt.Println("Tool: Draw Mode")
			case fyne.Key3, "KP3":
				mosugoCanvas.SetTool(tools.ToolErase)
				fmt.Println("Tool: Erase Mode")

			case fyne.KeyEscape:
				mosugoCanvas.SetTool(tools.ToolCard)
			}
			if targetBtn != nil {
				mosugoCanvas.SetTool(targetTool)
				for _, b := range toolButtons {
					b.Importance = widget.LowImportance
					b.Refresh()
				}
				targetBtn.Importance = widget.HighImportance
				targetBtn.Refresh()
			}
		})
	}

	w.SetContent(finalLayout)
	w.ShowAndRun()
}
