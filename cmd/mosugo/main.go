package main

import (
	"fmt"
	"image/color"
	"log"
	"time"

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
	w.Resize(fyne.NewSize(600, 500))
	w.SetPadded(false)

	if icon, err := fyne.LoadResourceFromPath("assets/Mosugo_Icon.png"); err == nil {
		w.SetIcon(icon)
	}

	mosugoCanvas := mosuCanvas.NewMosugoCanvas()

	var autoSaveTimer *time.Timer

	mosugoCanvas.SetOnDirty(func() {
		if autoSaveTimer != nil {
			autoSaveTimer.Stop()
		}
		autoSaveTimer = time.AfterFunc(2*time.Second, func() {
			err := mosugoCanvas.SaveCurrentWorkspace()
			if err != nil {
				log.Println("Auto-save failed:", err)
			} else {
				fmt.Println("Auto-saved workspace for", mosugoCanvas.GetCurrentDate().Format("2006-01-02"))
			}
		})
	})

	// Load today's workspace on startup
	today := time.Now()
	if err := mosugoCanvas.LoadWorkspace(today); err != nil {
		log.Println("Could not load today's workspace:", err)
	}

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
	metaBorder.SetCurrentDate(today)

	// Calendar content with date selection callback
	calendarContent := ui.NewCalendarContent(today, func(selectedDate time.Time) {
		// Save current workspace before switching
		if err := mosugoCanvas.SaveCurrentWorkspace(); err != nil {
			log.Println("Failed to save workspace:", err)
		}

		// Load selected date's workspace
		if err := mosugoCanvas.LoadWorkspace(selectedDate); err != nil {
			log.Println("Failed to load workspace for", selectedDate.Format("2006-01-02"), ":", err)
		} else {
			fmt.Println("Switched to workspace:", selectedDate.Format("2006-01-02"))
			metaBorder.SetCurrentDate(selectedDate)
			if metaBorder.BottomTabExpanded {
				metaBorder.ToggleCalendar()
			}
		}
	})

	metaBorder.SetCalendarContent(calendarContent)

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
			case fyne.Key0, "KP0":
				mosugoCanvas.SetTool(tools.ToolSelect)
				fmt.Println("Tool: Select Mode")
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

	// Add keyboard shortcuts using canvas shortcuts
	ctrlS := &desktop.CustomShortcut{KeyName: fyne.KeyS, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlS, func(shortcut fyne.Shortcut) {
		// Ctrl+S - Force save
		if err := mosugoCanvas.SaveCurrentWorkspace(); err != nil {
			log.Println("Manual save failed:", err)
		} else {
			fmt.Println("Workspace saved manually")
		}
	})

	ctrlLeft := &desktop.CustomShortcut{KeyName: fyne.KeyLeft, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlLeft, func(shortcut fyne.Shortcut) {
		// Ctrl+Left - Previous day
		currentDate := mosugoCanvas.GetCurrentDate()
		previousDay := currentDate.AddDate(0, 0, -1)

		if err := mosugoCanvas.SaveCurrentWorkspace(); err != nil {
			log.Println("Failed to save before navigating:", err)
		}
		if err := mosugoCanvas.LoadWorkspace(previousDay); err != nil {
			log.Println("Failed to load previous day:", err)
		} else {
			metaBorder.SetCurrentDate(previousDay)
		}

		fmt.Println("Navigated to:", previousDay.Format("2006-01-02"))
	})

	ctrlRight := &desktop.CustomShortcut{KeyName: fyne.KeyRight, Modifier: fyne.KeyModifierControl}
	w.Canvas().AddShortcut(ctrlRight, func(shortcut fyne.Shortcut) {
		currentDate := mosugoCanvas.GetCurrentDate()
		nextDay := currentDate.AddDate(0, 0, 1)

		if err := mosugoCanvas.SaveCurrentWorkspace(); err != nil {
			log.Println("Failed to save before navigating:", err)
		}
		if err := mosugoCanvas.LoadWorkspace(nextDay); err != nil {
			log.Println("Failed to load next day:", err)
		} else {
			metaBorder.SetCurrentDate(nextDay)
		}

		fmt.Println("Navigated to:", nextDay.Format("2006-01-02"))
	})

	w.SetContent(finalLayout)
	w.ShowAndRun()
}
