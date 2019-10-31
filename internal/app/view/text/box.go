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
	color [4]float32

	// bubble is the bubble to render behind the text. Nil for none.
	bubble *rect.Bubble

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle

	// size is the size of the text that will be rendered which may not be all the text.
	size image.Point

	// pt is the location to render the text.
	pt image.Point

	// dirty is true if the text or bounds has changed and requires an update.
	dirty bool
}

// Option is an option to pass to New.
type Option func(c *Box)

// NewBox creates a new Box.
func NewBox(textRenderer *gfx.TextRenderer, text string, opts ...Option) *Box {
	b := &Box{
		textRenderer: textRenderer,
		text:         text,
		color:        [4]float32{1, 1, 1, 1},
		dirty:        true,
	}
	for _, o := range opts {
		o(b)
	}
	return b
}

// Color returns an option to set the text color.
func Color(color [4]float32) Option {
	return func(b *Box) {
		b.color = color
	}
}

// Bubble returns an option to configure the background bubble.
func Bubble(rounding, padding int) Option {
	return func(b *Box) {
		b.bubble = &rect.Bubble{
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
	if b.text == text {
		return
	}
	b.text = text
	b.dirty = true
}

// SetBounds sets the bounds with global coordinates to draw within.
func (b *Box) SetBounds(bounds image.Rectangle) {
	if b.bounds == bounds {
		return
	}
	b.bounds = bounds
	b.dirty = true
}

// ProcessInput processes the input.
func (b *Box) ProcessInput() {}

// Update updates the state by one frame and returns true if another update is needed for animation.
func (b *Box) Update() (dirty bool) {
	if !b.dirty {
		return false
	}

	b.size = b.textRenderer.Measure(b.text)

	totalWidth := b.size.X
	if b.bubble != nil {
		totalWidth += b.bubble.Padding
	}

	if totalWidth > b.bounds.Dx() {
		b.size.X = b.bounds.Dx()
		if b.bubble != nil {
			b.size.X -= b.bubble.Padding * 2
		}
	}

	b.pt = image.Point{
		X: b.bounds.Min.X + b.bounds.Dx()/2 - b.size.X/2,
		Y: b.bounds.Min.Y + b.bounds.Dy()/2 - b.size.Y/2,
	}

	b.dirty = false

	return false
}

// Render renders the current state to the screen.
func (b *Box) Render(fudge float32) {
	if b.text == "" {
		return
	}

	if b.bounds.Empty() {
		return
	}

	if b.bubble != nil {
		b.bubble.Render(b.pt, b.size)
	}

	b.textRenderer.Render(b.text, b.pt, b.color, gfx.TextRenderMaxWidth(b.size.X))
}
