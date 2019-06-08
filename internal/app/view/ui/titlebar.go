package ui

import (
	"github.com/go-gl/glfw/v3.2/glfw"

	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/btmura/ponzi2/internal/app/view/status"
	"github.com/btmura/ponzi2/internal/errors"
)

// titleBar renders the the title bar.
type titleBar struct {
	win  *glfw.Window
	text string
}

func newTitleBar(win *glfw.Window) *titleBar {
	return &titleBar{
		win:  win,
		text: appName,
	}
}

// SetData sets the Title's stock.
func (t *titleBar) SetData(data *chart.Data) error {
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
func (t *titleBar) Render(fudge float32) error {
	t.win.SetTitle(t.text)
	return nil
}
