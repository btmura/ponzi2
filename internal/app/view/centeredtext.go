package view

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
)

// centeredText draws text that is horizontally and vertically centered.
type centeredText struct {
	// textRenderer renders the text.
	textRenderer *gfx.TextRenderer

	// Text is the text to render.
	Text string

	// color is the color to render the text in.
	color [3]float32

	// bubbleSpec specifies the bubble to render behind the text. Nil for none.
	bubbleSpec *bubbleSpec
}

// centeredTextOpt is an option to pass to NewCenteredText.
type centeredTextOpt func(c *centeredText)

// newCenteredText creates a new CenteredText.
func newCenteredText(textRenderer *gfx.TextRenderer, text string, opts ...centeredTextOpt) *centeredText {
	c := &centeredText{
		textRenderer: textRenderer,
		Text:         text,
		color:        white,
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// centeredTextColor returns an option to set the text color.
func centeredTextColor(color [3]float32) centeredTextOpt {
	return func(c *centeredText) {
		c.color = color
	}
}

// centeredTextBubble returns an option to configure the background bubble.
func centeredTextBubble(rounding, padding int) centeredTextOpt {
	return func(c *centeredText) {
		c.bubbleSpec = &bubbleSpec{
			rounding: rounding,
			padding:  padding,
		}
	}
}

// Render renders the CenteredText.
func (c *centeredText) Render(r image.Rectangle) {
	if c.Text == "" {
		return
	}

	sz := c.textRenderer.Measure(c.Text)

	pt := image.Point{
		X: r.Min.X + r.Dx()/2 - sz.X/2,
		Y: r.Min.Y + r.Dy()/2 - sz.Y/2,
	}

	if c.bubbleSpec != nil {
		renderBubble(pt, sz, *c.bubbleSpec)
	}

	c.textRenderer.Render(c.Text, pt, c.color)
}
