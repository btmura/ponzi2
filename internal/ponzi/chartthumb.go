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
	thumbLoadingText     = NewCenteredText(thumbSymbolQuoteTextRenderer, "LOADING...")
)

// ChartThumb shows a thumbnail for a stock.
type ChartThumb struct {
	header             *ChartHeader
	lines              *ChartLines
	dailyStochastics   *ChartStochastics
	weeklyStochastics  *ChartStochastics
	thumbClickCallback func()
	loading            bool
}

// NewChartThumb creates a ChartThumb.
func NewChartThumb() *ChartThumb {
	return &ChartThumb{
		header: NewChartHeader(&ChartHeaderArgs{
			SymbolQuoteTextRenderer: thumbSymbolQuoteTextRenderer,
			QuoteFormatter:          thumbFormatQuote,
			Button2:                 NewButton(thumbRemoveButtonVAO),
			Rounding:                thumbChartRounding,
			Padding:                 thumbChartPadding,
		}),
		lines:             &ChartLines{},
		dailyStochastics:  &ChartStochastics{Interval: DailyInterval},
		weeklyStochastics: &ChartStochastics{Interval: WeeklyInterval},
		loading:           true,
	}
}

// Update updates the ChartThumb.
func (ch *ChartThumb) Update(u *ChartUpdate) {
	ch.loading = u.Loading
	ch.header.Update(u)
	ch.lines.Update(u.Stock)
	ch.dailyStochastics.Update(u.Stock)
	ch.weeklyStochastics.Update(u.Stock)
}

// Render renders the chart thumbnail.
func (ch *ChartThumb) Render(vc ViewContext) {
	// Render the border around the chart.
	renderRoundedRect(vc.Bounds, thumbChartRounding)

	// Render the header and the line below it.
	r, _, clicked := ch.header.Render(vc)
	rects := sliceRect(r, 0.5, 0.5)
	renderHorizDivider(rects[1], horizLine)

	if ch.loading {
		thumbLoadingText.Render(r)
		return
	}

	if !clicked && vc.LeftClickInBounds() {
		vc.ScheduleCallback(ch.thumbClickCallback)
	}

	renderHorizDivider(rects[0], horizLine)

	ch.dailyStochastics.Render(rects[1].Inset(thumbChartPadding))
	ch.weeklyStochastics.Render(rects[0].Inset(thumbChartPadding))
	ch.lines.Render(rects[1].Inset(thumbChartPadding))
	ch.lines.Render(rects[0].Inset(thumbChartPadding))
}

// SetRemoveButtonClickCallback sets the callback for when the remove button is clicked.
func (ch *ChartThumb) SetRemoveButtonClickCallback(cb func()) {
	ch.header.SetButton2ClickCallback(cb)
}

// SetThumbClickCallback sets the callback for when the thumbnail is clicked.
func (ch *ChartThumb) SetThumbClickCallback(cb func()) {
	ch.thumbClickCallback = cb
}

// Close frees the resources backing the chart thumbnail.
func (ch *ChartThumb) Close() {
	if ch.header != nil {
		ch.header.Close()
		ch.header = nil
	}
	if ch.lines != nil {
		ch.lines.Close()
		ch.lines = nil
	}
	if ch.dailyStochastics != nil {
		ch.dailyStochastics.Close()
		ch.dailyStochastics = nil
	}
	if ch.weeklyStochastics != nil {
		ch.weeklyStochastics.Close()
		ch.weeklyStochastics = nil
	}
	ch.thumbClickCallback = nil
}
