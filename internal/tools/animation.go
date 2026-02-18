package tools

import (
	"time"

	"mosugo/internal/cards"

	"fyne.io/fyne/v2"
)

func AnimateCardBounce(c Canvas, card *cards.MosuWidget) {
	targetSize := card.WorldSize
	card.WorldSize = fyne.NewSize(0, 0)
	c.Refresh()

	// Bounce logic: overshoot then settle in 200ms
	start := time.Now()
	duration := 400 * time.Millisecond

	// Simple spring/bounce curve
	// y = 1 - (1-x)^2 * cos(x*pi*3) or similar
	// Using a simpler approach: Scale 0 -> 1.1 -> 1.0

	go func() {
		ticker := time.NewTicker(16 * time.Millisecond) // ~60fps
		defer ticker.Stop()

		for {
			t := time.Since(start)
			if t > duration {
				card.WorldSize = targetSize
				c.Refresh()
				return
			}

			progress := float32(t.Milliseconds()) / float32(duration.Milliseconds())

			// Elastic out easing
			// p = 2^(-10*t) * sin((t*10 - 0.75) * c4) + 1
			// Simplified bounce:
			var scale float32
			if progress < 0.7 {
				// Go slightly over 1.0
				scale = progress * (1.1 / 0.7)
			} else {
				// Return to 1.0 from 1.1
				scale = 1.1 - ((progress - 0.7) * (0.1 / 0.3))
			}

			// Just use a simpler standard elastic/back out
			// c1 := 1.70158
			// c3 := c1 + 1
			// scale = 1 + c3 * (progress - 1)^3 + c1 * (progress - 1)^2
			// That's BackOut.

			// Let's stick to a manual easing for "bounce in scale"
			// 0 -> 1.1 (at 70%) -> 1.0 (at 100%)

			w := targetSize.Width * scale
			h := targetSize.Height * scale

			card.WorldSize = fyne.NewSize(w, h)
			c.Refresh()
			<-ticker.C
		}
	}()
}
