package rect

import "image"

// Bubble is a rounded rectangle with a fill and stroke color.
type Bubble struct {
	// Rounding is how rounded the corners of the bubble are.
	Rounding int

	// Padding is how much padding to add around the given bounds.
	Padding int
}

// Render renders the bubble.
func (b *Bubble) Render(pt, sz image.Point) {
	br := image.Rectangle{
		Min: pt,
		Max: pt.Add(sz),
	}
	br = br.Inset(-b.Padding)
	FillRoundedRect(br, b.Rounding)
	StrokeRoundedRect(br, b.Rounding)
}
