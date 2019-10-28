package rect

import "image"

// TODO(btmura): change from bubbleSpec type to bubble type
type BubbleSpec struct {
	// Rounding is how much rounding of the bubble's rounded rectangle.
	Rounding int

	// Padding is how much padding of the bubble's text.
	Padding int
}

func RenderBubble(pt, sz image.Point, bs BubbleSpec) {
	br := image.Rectangle{
		Min: pt,
		Max: pt.Add(sz),
	}
	br = br.Inset(-bs.Padding)
	FillRoundedRect(br, bs.Rounding)
	StrokeRoundedRect(br, bs.Rounding)
}
