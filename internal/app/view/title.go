package view

import (
	"github.com/go-gl/glfw/v3.2/glfw"

	"gitlab.com/btmura/ponzi2/internal/status"
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
func (t *Title) SetData(data *ChartData) error {
	if data == nil {
		return status.Error("missing data")
	}

	q := data.Quote

	if q == nil {
		t.text = join(data.Symbol, "-", appName)
		return nil
	}

	t.text = join(data.Symbol, paren(q.CompanyName), priceStatus(q), updateStatus(q), "-", appName)

	return nil
}

// Render renders the Title.
func (t *Title) Render(win *glfw.Window) {
	win.SetTitle(t.text)
}
