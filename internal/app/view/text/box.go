package text

import (
	"image"

	"github.com/btmura/ponzi2/internal/app/gfx"
	"github.com/btmura/ponzi2/internal/app/view/color"
	"github.com/btmura/ponzi2/internal/app/view/rect"
)

// Box draws text that is horizontally and vertically centered.
type Box struct {
	// textRenderer renders the text.
	textRenderer *gfx.TextRenderer

	// text is the text to render.
	text string

	// color is the color to render the text in.
	color color.RGBA

	// padding is the padding around the text.
	padding int

	// bubble is the bubble around the text. Nil for no bubble.
	bubble *rect.Bubble

	// bounds is the rectangle with global coords that should be drawn within.
	bounds image.Rectangle

	// textBounds is the adjusted bounds shrink-wrapped around the text.
	textBounds image.Rectangle

	// dirty is true if the text or bounds has changed and requires an update.
	dirty bool
}

// Option is an option to pass to New.
type Option func(c *Box)

// Color returns an option to set the text color.
func Color(color color.RGBA) Option {
	return func(b *Box) {
		b.color = color
	}
}

// Padding returns an option to set the padding.
func Padding(padding int) Option {
	return func(b *Box) {
		b.padding = padding
	}
}

// Bubble returns an option to set the bubble.
func Bubble(bubble *rect.Bubble) Option {
	return func(b *Box) {
		b.bubble = bubble
	}
}

// NewBox creates a new Box.
func NewBox(textRenderer *gfx.TextRenderer, text string, opts ...Option) *Box {
	b := &Box{
		textRenderer: textRenderer,
		text:         text,
		color:        color.RGBA{1, 1, 1, 1},
		dirty:        true,
	}
	for _, o := range opts {
		o(b)
	}
	return b
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

// Update updates the state by one frame and returns true if another update is needed for animation.
func (b *Box) Update() (dirty bool) {
	if !b.dirty {
		return false
	}

	// Figure out the width of the text. Trim to fit within bounds.
	textSize := b.textRenderer.Measure(b.text)
	if textSize.X > b.bounds.Dx()-b.padding {
		textSize.X = b.bounds.Dx() - b.padding
	}

	// Figure out the shrink-wrapped bounds of the text.
	textPt := image.Point{
		X: b.bounds.Min.X + b.bounds.Dx()/2 - textSize.X/2,
		Y: b.bounds.Min.Y + b.bounds.Dy()/2 - textSize.Y/2,
	}
	b.textBounds = image.Rectangle{
		Min: textPt,
		Max: textPt.Add(textSize),
	}

	if b.bubble != nil {
		b.bubble.SetBounds(b.textBounds.Inset(-b.padding))
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
		b.bubble.Render(fudge)
	}

	b.textRenderer.Render(b.text, b.textBounds.Min, gfx.TextColor(b.color), gfx.TextRenderMaxWidth(b.textBounds.Dx()))
}
