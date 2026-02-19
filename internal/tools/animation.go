package tools

import (
	"time"

	"mosugo/internal/cards"

	"fyne.io/fyne/v2"
)

func BounceEasing(t float32) float32 {
	if t < 0.4 {
		return (t / 0.4) * 1.2
	} else if t < 0.7 {
		return 1.2 - ((t - 0.4) / 0.3) * 0.4
	} else {
		return 0.8 + ((t - 0.7) / 0.3) * 0.2
	}
}

func AnimateCardBounce(c Canvas, card *cards.MosuWidget) {
	targetSize := card.WorldSize
	targetPos := card.WorldPos
	
	centerX := targetPos.X + targetSize.Width/2
	centerY := targetPos.Y + targetSize.Height/2
	
	card.WorldSize = fyne.NewSize(0, 0)
	card.WorldPos = fyne.NewPos(centerX, centerY)
	c.Refresh()

	start := time.Now()
	duration := 400 * time.Millisecond

	go func() {
		ticker := time.NewTicker(16 * time.Millisecond)
		defer ticker.Stop()

		for {
			t := time.Since(start)
			if t > duration {
				card.WorldSize = targetSize
				card.WorldPos = targetPos
				c.Refresh()
				return
			}

			progress := float32(t.Milliseconds()) / float32(duration.Milliseconds())
			scale := BounceEasing(progress)

			w := targetSize.Width * scale
			h := targetSize.Height * scale
			
			x := centerX - w/2
			y := centerY - h/2

			card.WorldSize = fyne.NewSize(w, h)
			card.WorldPos = fyne.NewPos(x, y)
			c.Refresh()
			<-ticker.C
		}
	}()
}
