package app

import (
	"image"

	"github.com/btmura/ponzi2/internal/gfx"
)

// CenteredText draws text that is horizontally and vertically centered.
type CenteredText struct {
	// textRenderer renders the text.
	textRenderer *gfx.TextRenderer

	// Text is the text to render.
	Text string

	// color is the color to render the text in.
	color [3]float32

	// bubbleSpec specifies the bubble to render behind the text. Nil for none.
	bubbleSpec *centeredTextBubbleSpec
}

type centeredTextBubbleSpec struct {
	// rounding is how much rounding of the bubble's rounded rectangle.
	rounding int

	// padding is how much padding of the bubble's text.
	padding int
}

// CenteredTextOpt is an option to pass to NewCenteredText.
type CenteredTextOpt func(c *CenteredText)

// NewCenteredText creates a new CenteredText.
func NewCenteredText(textRenderer *gfx.TextRenderer, text string, opts ...CenteredTextOpt) *CenteredText {
	c := &CenteredText{
		textRenderer: textRenderer,
		Text:         text,
		color:        white,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// CenteredTextColor returns an option to set the text color.
func CenteredTextColor(color [3]float32) CenteredTextOpt {
	return func(c *CenteredText) {
		c.color = color
	}
}

// CenteredTextBubble returns an option to configure the background bubble.
func CenteredTextBubble(rounding, padding int) CenteredTextOpt {
	return func(c *CenteredText) {
		c.bubbleSpec = &centeredTextBubbleSpec{
			rounding: rounding,
			padding:  padding,
		}
	}
}

// Render renders the CenteredText.
func (c *CenteredText) Render(r image.Rectangle) {
	if c.Text == "" {
		return
	}

	sz := c.textRenderer.Measure(c.Text)

	pt := image.Point{
		X: r.Min.X + r.Dx()/2 - sz.X/2,
		Y: r.Min.Y + r.Dy()/2 - sz.Y/2,
	}

	if bs := c.bubbleSpec; bs != nil {
		br := image.Rectangle{
			Min: pt,
			Max: pt.Add(sz),
		}
		br = br.Inset(-bs.padding)

		fillRoundedRect(br, bs.rounding)
		strokeRoundedRect(br, bs.rounding)
	}

	c.textRenderer.Render(c.Text, pt, c.color)
}
