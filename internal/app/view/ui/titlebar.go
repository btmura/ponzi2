package ui

import (
	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/go-gl/glfw/v3.3/glfw"

	"github.com/btmura/ponzi2/internal/app/view/status"
)

// titleBar renders the the title bar.
type titleBar struct {
	win  *glfw.Window
	text string
}

// newTitleBar returns a new title bar.
func newTitleBar(win *glfw.Window) *titleBar {
	return &titleBar{
		win:  win,
		text: appName,
	}
}

// SetData sets the title bar's stock.
func (t *titleBar) SetData(data chart.Data) {
	q := data.Quote

	if q == nil {
		t.text = status.Join(data.Symbol, "-", appName)
		return
	}

	t.text = status.Join(data.Symbol, status.Paren(q.CompanyName), status.PriceChange(q), status.SourceUpdate(q), "-", appName)
}

// Render renders the title bar.
func (t *titleBar) Render(float32) {
	t.win.SetTitle(t.text)
}
