package ponzi

import (
	"bytes"
	"fmt"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/gfx"
)

const (
	chartRounding = 10
	chartPadding  = 5
)

var (
	chartSymbolQuoteTextRenderer = gfx.NewTextRenderer(goregular.TTF, 24)
	chartFormatQuote             = func(st *ModelStock) string {
		if st.Price() != 0 {
			return fmt.Sprintf("%.2f %+5.2f %+5.2f%%", st.Price(), st.Change(), st.PercentChange()*100.0)
		}
		return ""
	}
	chartAddButtonVAO = gfx.ReadPLYVAO(bytes.NewReader(MustAsset("addButton.ply")))
)

// Shared variables used by multiple chart components.
var (
	chartAxisLabelTextRenderer = gfx.NewTextRenderer(goregular.TTF, 12)
	chartGridHorizLine         = gfx.HorizColoredLineVAO(gray, gray)
)

// Chart shows a stock chart for a single stock.
type Chart struct {
	stock                   *ModelStock
	header                  *ChartHeader
	lines                   *ChartLines
	prices                  *ChartPrices
	volume                  *ChartVolume
	dailyStochastics        *ChartStochastics
	weeklyStochastics       *ChartStochastics
	addButtonClickCallbacks []func()
}

// NewChart creates a new chart.
func NewChart(stock *ModelStock) *Chart {
	return &Chart{
		stock:             stock,
		header:            NewChartHeader(stock, chartSymbolQuoteTextRenderer, chartFormatQuote, NewButton(chartAddButtonVAO), chartRounding, chartPadding),
		lines:             NewChartLines(stock),
		prices:            NewChartPrices(stock),
		volume:            NewChartVolume(stock),
		dailyStochastics:  NewChartStochastics(stock, DailySTO),
		weeklyStochastics: NewChartStochastics(stock, WeeklySTO),
	}
}

// Update updates the chart.
func (ch *Chart) Update() {
	ch.lines.Update()
	ch.prices.Update()
	ch.volume.Update()
	ch.dailyStochastics.Update()
	ch.weeklyStochastics.Update()
}

// Render renders the chart.
func (ch *Chart) Render(vc ViewContext) {
	r, _ := ch.header.Render(vc)

	rects := renderHorizDividers(r, horizLine, 0.13, 0.13, 0.13, 0.61)
	pr, vr, dr, wr := rects[3], rects[2], rects[1], rects[0]

	pr = pr.Inset(chartPadding)
	vr = vr.Inset(chartPadding)
	dr = dr.Inset(chartPadding)
	wr = wr.Inset(chartPadding)

	maxWidth := ch.prices.RenderLabels(pr)
	if w := ch.volume.RenderLabels(vr); w > maxWidth {
		maxWidth = w
	}
	if w := ch.dailyStochastics.RenderLabels(dr); w > maxWidth {
		maxWidth = w
	}
	if w := ch.weeklyStochastics.RenderLabels(wr); w > maxWidth {
		maxWidth = w
	}

	pr.Max.X -= maxWidth + chartPadding
	vr.Max.X -= maxWidth + chartPadding
	dr.Max.X -= maxWidth + chartPadding
	wr.Max.X -= maxWidth + chartPadding

	ch.prices.Render(pr)
	ch.volume.Render(vr)
	ch.dailyStochastics.Render(dr)
	ch.weeklyStochastics.Render(wr)

	ch.lines.Render(pr)
	ch.lines.Render(vr)
	ch.lines.Render(dr)
	ch.lines.Render(wr)
}

// AddAddButtonClickCallback adds a callback for when the add button is clicked.
func (ch *Chart) AddAddButtonClickCallback(cb func()) {
	ch.header.AddButtonClickCallback(cb)
}

// Close frees the resources backing the chart.
func (ch *Chart) Close() {
	ch.prices.Close()
	ch.prices = nil
	ch.volume.Close()
	ch.volume = nil
	ch.dailyStochastics.Close()
	ch.dailyStochastics = nil
	ch.weeklyStochastics.Close()
	ch.weeklyStochastics = nil
}
