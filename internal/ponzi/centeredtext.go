package ponzi

import (
	"github.com/btmura/ponzi2/internal/gfx"
)

// CenteredText draws text that is horizontally and vertically centered.
type CenteredText struct {
	// Text is the text to render.
	Text string

	// textRenderer renders the text.
	textRenderer *gfx.TextRenderer
}

// NewCenteredText creates a new CenteredText.
func NewCenteredText(textRenderer *gfx.TextRenderer) *CenteredText {
	return &CenteredText{
		textRenderer: textRenderer,
	}
}

// Render renders the CenteredText.
func (c *CenteredText) Render(vc ViewContext) {
	sz := c.textRenderer.Measure(c.Text)
	pt := vc.Bounds.Min
	pt.X += (vc.Bounds.Dx() - sz.X) / 2
	pt.Y += (vc.Bounds.Dy() - sz.Y) / 2
	c.textRenderer.Render(c.Text, pt, white)
}
