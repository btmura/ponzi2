package view

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/view/rect"
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
	bubbleSpec *rect.BubbleSpec

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
		c.bubbleSpec = &rect.BubbleSpec{
			Rounding: rounding,
			Padding:  padding,
		}
	}
}

func (c *centeredText) ProcessInput(bounds image.Rectangle) {
	c.size = c.textRenderer.Measure(c.Text)
	c.pt = image.Point{
		X: bounds.Min.X + bounds.Dx()/2 - c.size.X/2,
		Y: bounds.Min.Y + bounds.Dy()/2 - c.size.Y/2,
	}
}

// Render renders the CenteredText.
func (c *centeredText) Render(fudge float32) {
	if c.Text == "" {
		return
	}

	if c.bubbleSpec != nil {
		rect.RenderBubble(c.pt, c.size, *c.bubbleSpec)
	}

	c.textRenderer.Render(c.Text, c.pt, c.color)
}
