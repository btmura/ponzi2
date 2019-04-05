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

	// size is the measured size of the rendered text.
	size image.Point

	// pt is the location to render the text.
	pt image.Point
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

func (c *centeredText) ProcessInput(ic inputContext) {
	c.size = c.textRenderer.Measure(c.Text)

	r := ic.Bounds

	c.pt = image.Point{
		X: r.Min.X + r.Dx()/2 - c.size.X/2,
		Y: r.Min.Y + r.Dy()/2 - c.size.Y/2,
	}
}

// Render renders the CenteredText.
func (c *centeredText) Render(fudge float32) {
	if c.Text == "" {
		return
	}

	if c.bubbleSpec != nil {
		renderBubble(c.pt, c.size, *c.bubbleSpec)
	}

	c.textRenderer.Render(c.Text, c.pt, c.color)
}
