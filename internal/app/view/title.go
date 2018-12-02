package view

import (
	"github.com/go-gl/glfw/v3.2/glfw"

	"gitlab.com/btmura/ponzi2/internal/app/model"
)

// Title renders the the title bar.
type Title struct {
	text string
}

// NewTitle creates a new Title.
func NewTitle() *Title {
	return &Title{text: appName}
}

// SetData sets the Title's stock.
func (t *Title) SetData(st *model.Stock) {
	q := st.Quote
	if q == nil {
		t.text = join(st.Symbol, "-", appName)
		return
	}
	t.text = join(st.Symbol, paren(q.CompanyName), priceStatus(st.Quote), updateStatus(st.Quote), "-", appName)
}

// Render renders the Title.
func (t *Title) Render(win *glfw.Window) {
	win.SetTitle(t.text)
}
