package rect

import "image"

// Bubble is a rounded rectangle with a fill and stroke color.
type Bubble struct {
	// Rounding is how rounded the corners of the bubble are.
	Rounding int
}

// Render renders the bubble.
func (b *Bubble) Render(bounds image.Rectangle) {
	FillRoundedRect(bounds, b.Rounding)
	StrokeRoundedRect(bounds, b.Rounding)
}
