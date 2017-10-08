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
	thumbLoadingText     = NewCenteredText(thumbSymbolQuoteTextRenderer, "LOADING")
)

// ChartThumbnail shows a stock chart thumbnail for a single stock.
type ChartThumbnail struct {
	Loading            bool
	header             *ChartHeader
	lines              *ChartLines
	dailyStochastics   *ChartStochastics
	weeklyStochastics  *ChartStochastics
	thumbClickCallback func()
}

// NewChartThumbnail creates a ChartThumbnail.
func NewChartThumbnail() *ChartThumbnail {
	return &ChartThumbnail{
		Loading:           true,
		header:            NewChartHeader(thumbSymbolQuoteTextRenderer, thumbFormatQuote, NewButton(thumbRemoveButtonVAO), thumbChartRounding, thumbChartPadding),
		lines:             &ChartLines{},
		dailyStochastics:  &ChartStochastics{Interval: DailyInterval},
		weeklyStochastics: &ChartStochastics{Interval: WeeklyInterval},
	}
}

// Update updates the ChartThumbnail.
func (ch *ChartThumbnail) Update(st *ModelStock) {
	ch.header.Update(st)
	ch.lines.Update(st)
	ch.dailyStochastics.Update(st)
	ch.weeklyStochastics.Update(st)
}

// Render renders the chart thumbnail.
func (ch *ChartThumbnail) Render(vc ViewContext) {
	if ch.Loading {
		thumbLoadingText.Render(vc)
		return
	}

	r, clicked := ch.header.Render(vc)

	if !clicked && vc.LeftClickInBounds() {
		vc.ScheduleCallback(ch.thumbClickCallback)
	}

	rects := renderHorizDividers(r, horizLine, 0.5, 0.5)

	ch.dailyStochastics.Render(rects[1].Inset(thumbChartPadding))
	ch.weeklyStochastics.Render(rects[0].Inset(thumbChartPadding))
	ch.lines.Render(rects[1].Inset(thumbChartPadding))
	ch.lines.Render(rects[0].Inset(thumbChartPadding))
}

// SetRemoveButtonClickCallback sets the callback for when the remove button is clicked.
func (ch *ChartThumbnail) SetRemoveButtonClickCallback(cb func()) {
	ch.header.SetButtonClickCallback(cb)
}

// SetThumbClickCallback sets the callback for when the thumbnail is clicked.
func (ch *ChartThumbnail) SetThumbClickCallback(cb func()) {
	ch.thumbClickCallback = cb
}

// Close frees the resources backing the chart thumbnail.
func (ch *ChartThumbnail) Close() {
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
