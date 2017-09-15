package ponzi

import (
	"bytes"
	"fmt"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/gfx"
)

const (
	thumbChartRounding = 6
	thumbChartPadding  = 2
)

var (
	thumbSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 12)
	thumbFormatQuote             = func(st *ModelStock) string {
		if st.Price() != 0 {
			return fmt.Sprintf(" %.2f %+5.2f%% ", st.Price(), st.PercentChange()*100.0)
		}
		return ""
	}
	thumbRemoveButtonVAO = gfx.ReadPLYVAO(bytes.NewReader(MustAsset("removeButton.ply")))
)

// ChartThumbnail shows a stock chart thumbnail for a single stock.
type ChartThumbnail struct {
	stock                      *ModelStock
	header                     *ChartHeader
	lines                      *ChartLines
	dailyStochastics           *ChartStochastics
	weeklyStochastics          *ChartStochastics
	removeButtonClickCallbacks []func()
	thumbClickCallbacks        []func()
}

// NewChartThumbnail creates a new chart thumbnail.
func NewChartThumbnail(stock *ModelStock) *ChartThumbnail {
	return &ChartThumbnail{
		stock:             stock,
		header:            NewChartHeader(stock, thumbSymbolQuoteTextRenderer, thumbFormatQuote, NewButton(thumbRemoveButtonVAO), thumbChartRounding, thumbChartPadding),
		lines:             NewChartLines(stock),
		dailyStochastics:  NewChartStochastics(stock, DailySTO),
		weeklyStochastics: NewChartStochastics(stock, WeeklySTO),
	}
}

// Update updates the chart thumbnail.
func (ch *ChartThumbnail) Update() {
	ch.lines.Update()
	ch.dailyStochastics.Update()
	ch.weeklyStochastics.Update()
}

// Render renders the chart thumbnail.
func (ch *ChartThumbnail) Render(vc ViewContext) {
	r, clicked := ch.header.Render(vc)

	if !clicked && vc.LeftClickInBounds() {
		vc.ScheduleCallbacks(ch.thumbClickCallbacks)
	}

	rects := renderHorizDividers(r, horizLine, 0.5, 0.5)

	ch.dailyStochastics.Render(rects[1].Inset(thumbChartPadding))
	ch.weeklyStochastics.Render(rects[0].Inset(thumbChartPadding))
	ch.lines.Render(rects[1].Inset(thumbChartPadding))
	ch.lines.Render(rects[0].Inset(thumbChartPadding))
}

// AddRemoveButtonClickCallback adds a callback for when the remove button is clicked.
func (ch *ChartThumbnail) AddRemoveButtonClickCallback(cb func()) {
	ch.header.AddButtonClickCallback(cb)
}

// AddThumbClickCallback adds a callback for when the thumbnail is clicked
func (ch *ChartThumbnail) AddThumbClickCallback(cb func()) {
	ch.thumbClickCallbacks = append(ch.thumbClickCallbacks, cb)
}

// Close frees the resources backing the chart thumbnail.
func (ch *ChartThumbnail) Close() {
	ch.lines.Close()
	ch.lines = nil
	ch.dailyStochastics.Close()
	ch.dailyStochastics = nil
	ch.weeklyStochastics.Close()
	ch.weeklyStochastics = nil
}
