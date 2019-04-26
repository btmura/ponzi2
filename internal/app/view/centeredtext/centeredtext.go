package centeredtext

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/view/rect"
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
	bubbleSpec *rect.BubbleSpec

	// size is the measured size of the rendered text.
	size image.Point

	// pt is the location to render the text.
	pt image.Point
}

// Option is an option to pass to New.
type Option func(c *CenteredText)

// New creates a new CenteredText.
func New(textRenderer *gfx.TextRenderer, text string, opts ...Option) *CenteredText {
	c := &CenteredText{
		textRenderer: textRenderer,
		Text:         text,
		color:        [3]float32{1, 1, 1},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Color returns an option to set the text color.
func Color(color [3]float32) Option {
	return func(c *CenteredText) {
		c.color = color
	}
}

// Bubble returns an option to configure the background bubble.
func Bubble(rounding, padding int) Option {
	return func(c *CenteredText) {
		c.bubbleSpec = &rect.BubbleSpec{
			Rounding: rounding,
			Padding:  padding,
		}
	}
}

func (c *CenteredText) ProcessInput(bounds image.Rectangle) {
	c.size = c.textRenderer.Measure(c.Text)
	c.pt = image.Point{
		X: bounds.Min.X + bounds.Dx()/2 - c.size.X/2,
		Y: bounds.Min.Y + bounds.Dy()/2 - c.size.Y/2,
	}
}

// Render renders the CenteredText.
func (c *CenteredText) Render(fudge float32) {
	if c.Text == "" {
		return
	}

	if c.bubbleSpec != nil {
		rect.RenderBubble(c.pt, c.size, *c.bubbleSpec)
	}

	c.textRenderer.Render(c.Text, c.pt, c.color)
}
