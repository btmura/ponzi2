package title

import (
	"github.com/go-gl/glfw/v3.2/glfw"

	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/btmura/ponzi2/internal/app/view/status"
	"github.com/btmura/ponzi2/internal/errors"
)

// Application name for the window title.
// TODO(btmura): remove duplication with view.go
const appName = "ponzi2"

// Title renders the the title bar.
type Title struct {
	text string
}

// New creates a new Title.
func New() *Title {
	return &Title{text: appName}
}

// SetData sets the Title's stock.
func (t *Title) SetData(data *chart.Data) error {
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
