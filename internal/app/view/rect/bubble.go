package rect

import "image"

// TODO(btmura): fix bubble to not render outside the given bounds

// Bubble is a rounded rectangle with a fill and stroke color.
type Bubble struct {
	// Rounding is how rounded the corners of the bubble are.
	Rounding int

	// Padding is how much padding to add around the given bounds.
	Padding int
}

// Render renders the bubble.
func (b *Bubble) Render(bounds image.Rectangle) {
	bounds = bounds.Inset(-b.Padding)
	FillRoundedRect(bounds, b.Rounding)
	StrokeRoundedRect(bounds, b.Rounding)
}
