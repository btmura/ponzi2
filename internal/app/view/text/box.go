package text

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/view/rect"
)

// Box draws text that is horizontally and vertically centered.
type Box struct {
	// textRenderer renders the text.
	textRenderer *gfx.TextRenderer

	// text is the text to render.
	text string

	// color is the color to render the text in.
	color [3]float32

	// bubbleSpec specifies the bubble to render behind the text. Nil for none.
	bubbleSpec *rect.BubbleSpec

	// size is the measured size of the rendered text.
	size image.Point

	// pt is the location to render the text.
	pt image.Point

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle
}

// Option is an option to pass to New.
type Option func(c *Box)

// NewBox creates a new Box.
func NewBox(textRenderer *gfx.TextRenderer, text string, opts ...Option) *Box {
	b := &Box{
		textRenderer: textRenderer,
		text:         text,
		color:        [3]float32{1, 1, 1},
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

// Color returns an option to set the text color.
func Color(color [3]float32) Option {
	return func(b *Box) {
		b.color = color
	}
}

// Bubble returns an option to configure the background bubble.
func Bubble(rounding, padding int) Option {
	return func(b *Box) {
		b.bubbleSpec = &rect.BubbleSpec{
			Rounding: rounding,
			Padding:  padding,
		}
	}
}

// Text returns the text that will be shown in the box.
func (b *Box) Text() string {
	return b.text
}

// SetText sets the text to show in the box.
func (b *Box) SetText(text string) {
	b.text = text
}

// SetBounds sets the bounds with global coordinates to draw within.
func (b *Box) SetBounds(bounds image.Rectangle) {
	b.bounds = bounds
}

// ProcessInput processes the input.
func (b *Box) ProcessInput() {
	b.size = b.textRenderer.Measure(b.text)
	b.pt = image.Point{
		X: b.bounds.Min.X + b.bounds.Dx()/2 - b.size.X/2,
		Y: b.bounds.Min.Y + b.bounds.Dy()/2 - b.size.Y/2,
	}
}

// Render renders the Box.
func (b *Box) Render(fudge float32) {
	if b.text == "" {
		return
	}

	if b.bubbleSpec != nil {
		rect.RenderBubble(b.pt, b.size, *b.bubbleSpec)
	}

	b.textRenderer.Render(b.text, b.pt, b.color)
}
