package view

import (
	"github.com/go-gl/glfw/v3.2/glfw"

	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/btmura/ponzi2/internal/app/view/status"
	"github.com/btmura/ponzi2/internal/errors"
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
func (t *Title) SetData(data *chart.ChartData) error {
	if data == nil {
		return errors.Errorf("missing data")
	}

	q := data.Quote

	if q == nil {
		t.text = status.Join(data.Symbol, "-", appName)
		return nil
	}

	t.text = status.Join(data.Symbol, status.Paren(q.CompanyName), status.PriceChange(q), status.SourceUpdate(q), "-", appName)

	return nil
}

// Render renders the Title.
func (t *Title) Render(win *glfw.Window) {
	win.SetTitle(t.text)
}
