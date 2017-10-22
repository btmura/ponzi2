package ponzi

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
}

// NewCenteredText creates a new CenteredText.
func NewCenteredText(textRenderer *gfx.TextRenderer, text string, color [3]float32) *CenteredText {
	return &CenteredText{
		textRenderer: textRenderer,
		Text:         text,
		color:        color,
	}
}

// Render renders the CenteredText.
func (c *CenteredText) Render(r image.Rectangle) {
	if c.Text == "" {
		return
	}
	sz := c.textRenderer.Measure(c.Text)
	pt := r.Min
	pt.X += (r.Dx() - sz.X) / 2
	pt.Y += (r.Dy() - sz.Y) / 2
	c.textRenderer.Render(c.Text, pt, c.color)
}
