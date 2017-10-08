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
	chartLoadingText  = NewCenteredText(chartSymbolQuoteTextRenderer, "LOADING...")
)

// Shared variables used by multiple chart components.
var (
	chartAxisLabelTextRenderer = gfx.NewTextRenderer(goregular.TTF, 12)
	chartGridHorizLine         = gfx.HorizColoredLineVAO(gray, gray)
)

// Chart shows a stock chart for a single stock.
type Chart struct {
	header            *ChartHeader
	lines             *ChartLines
	prices            *ChartPrices
	volume            *ChartVolume
	dailyStochastics  *ChartStochastics
	weeklyStochastics *ChartStochastics
	loading           bool
}

// NewChart creates a new chart.
func NewChart() *Chart {
	return &Chart{
		header: NewChartHeader(&ChartHeaderArgs{
			SymbolQuoteTextRenderer: chartSymbolQuoteTextRenderer,
			QuoteFormatter:          chartFormatQuote,
			Button:                  NewButton(chartAddButtonVAO),
			Rounding:                chartRounding,
			Padding:                 chartPadding,
		}),
		lines:             &ChartLines{},
		prices:            &ChartPrices{},
		volume:            &ChartVolume{},
		dailyStochastics:  &ChartStochastics{Interval: DailyInterval},
		weeklyStochastics: &ChartStochastics{Interval: WeeklyInterval},
		loading:           true,
	}
}

// ChartUpdate is the argument to Update.
type ChartUpdate struct {
	// Loading indicates whether data is being fetched for the chart.
	Loading bool

	// Stock is the stock the chart should show.
	Stock *ModelStock
}

// Update updates the Chart.
func (ch *Chart) Update(u *ChartUpdate) {
	ch.loading = u.Loading
	ch.header.Update(u)
	ch.lines.Update(u.Stock)
	ch.prices.Update(u.Stock)
	ch.volume.Update(u.Stock)
	ch.dailyStochastics.Update(u.Stock)
	ch.weeklyStochastics.Update(u.Stock)
}

// Render renders the chart.
func (ch *Chart) Render(vc ViewContext) {
	// Render the border around the chart.
	renderRoundedRect(vc.Bounds, chartRounding)

	r, _ := ch.header.Render(vc)

	if ch.loading {
		vc := vc
		vc.Bounds = r
		renderHorizDividers(r, horizLine, 1)
		chartLoadingText.Render(vc)
		return
	}

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

// SetAddButtonClickCallback sets the callback for when the add button is clicked.
func (ch *Chart) SetAddButtonClickCallback(cb func()) {
	ch.header.SetButtonClickCallback(cb)
}

// Close frees the resources backing the chart.
func (ch *Chart) Close() {
	if ch.header != nil {
		ch.header.Close()
		ch.header = nil
	}
	if ch.lines != nil {
		ch.lines.Close()
		ch.lines = nil
	}
	if ch.prices != nil {
		ch.prices.Close()
		ch.prices = nil
	}
	if ch.volume != nil {
		ch.volume.Close()
		ch.volume = nil
	}
	if ch.dailyStochastics != nil {
		ch.dailyStochastics.Close()
		ch.dailyStochastics = nil
	}
	if ch.weeklyStochastics != nil {
		ch.weeklyStochastics.Close()
		ch.weeklyStochastics = nil
	}
}
