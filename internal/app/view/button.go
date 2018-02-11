package view

import (
	math "math"

	"github.com/btmura/ponzi2/internal/gfx"
)

const (
	buttonSpinTimeSec = 0.5
	buttonSpinFrames  = buttonSpinTimeSec * fps
	buttonSpinRadians = -2 * math.Pi / buttonSpinFrames
)

// Button is a button that can be rendered and clicked.
type Button struct {
	// iconVAO is the VAO to render.
	iconVAO *gfx.VAO

	// clickCallback is the callback to schedule when the button is clicked.
	clickCallback func()

	// spinning indicates whether to keep rotating the button.
	spinning bool

	// spinFrameIndex is the index into the animation. Used to finish a spin.
	spinFrameIndex int
}

// NewButton creates a new button.
func NewButton(iconVAO *gfx.VAO) *Button {
	return &Button{
		iconVAO: iconVAO,
	}
}

// StartSpinning starts the spinning animation.
func (b *Button) StartSpinning() {
	b.spinning = true
}

// StopSpinning stops the spinning animation.
func (b *Button) StopSpinning() {
	b.spinning = false
}

// Update updates the Button.
func (b *Button) Update() {
	if b.spinning || b.spinFrameIndex != 0 {
		b.spinFrameIndex = (b.spinFrameIndex + 1) % buttonSpinFrames
	}
}

// Render renders the Button and detects clicks.
func (b *Button) Render(vc ViewContext) (clicked bool) {
	if vc.LeftClickInBounds() {
		*vc.ScheduledCallbacks = append(*vc.ScheduledCallbacks, b.clickCallback)
		clicked = true
	}

	var spinRadians float32
	if b.spinFrameIndex > 0 {
		spinRadians = (float32(b.spinFrameIndex) + vc.Fudge) * buttonSpinRadians
	}
	gfx.SetModelMatrixRotatedRect(vc.Bounds, spinRadians)
	b.iconVAO.Render()

	return clicked
}

// SetClickCallback sets the callback for when the button is clicked.
func (b *Button) SetClickCallback(cb func()) {
	b.clickCallback = cb
}
