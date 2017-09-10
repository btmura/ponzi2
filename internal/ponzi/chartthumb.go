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
	thumbFormatQuote             = func(q *ModelQuote) string {
		if q.price != 0 {
			return fmt.Sprintf(" %.2f %+5.2f%% ", q.price, q.percentChange*100.0)
		}
		return ""
	}
	thumbRemoveButtonVAO = gfx.ReadPLYVAO(bytes.NewReader(MustAsset("removeButton.ply")))
)

type ChartThumbnail struct {
	stock                      *ModelStock
	header                     *ChartHeader
	lines                      *ChartLines
	dailyStochastics           *ChartStochastics
	weeklyStochastics          *ChartStochastics
	removeButtonClickCallbacks []func()
	thumbClickCallbacks        []func()
}

func NewChartThumbnail(stock *ModelStock) *ChartThumbnail {
	return &ChartThumbnail{
		stock:             stock,
		header:            NewChartHeader(stock, thumbSymbolQuoteTextRenderer, thumbFormatQuote, NewButton(thumbRemoveButtonVAO), thumbChartRounding, thumbChartPadding),
		lines:             NewChartLines(stock),
		dailyStochastics:  NewChartStochastics(stock, DailySTO),
		weeklyStochastics: NewChartStochastics(stock, WeeklySTO),
	}
}

func (ch *ChartThumbnail) Update() {
	ch.lines.Update()
	ch.dailyStochastics.Update()
	ch.weeklyStochastics.Update()
}

func (ch *ChartThumbnail) Render(vc ViewContext) {
	r, clicked := ch.header.Render(vc)

	if !clicked && vc.LeftClickInBounds() {
		vc.scheduleCallbacks(ch.thumbClickCallbacks)
	}

	rects := renderHorizDividers(r, horizLine, 0.5, 0.5)

	ch.dailyStochastics.Render(rects[1].Inset(thumbChartPadding))
	ch.weeklyStochastics.Render(rects[0].Inset(thumbChartPadding))
	ch.lines.Render(rects[1].Inset(thumbChartPadding))
	ch.lines.Render(rects[0].Inset(thumbChartPadding))
}

func (ch *ChartThumbnail) AddRemoveButtonClickCallback(cb func()) {
	ch.header.AddButtonClickCallback(cb)
}

func (ch *ChartThumbnail) AddThumbClickCallback(cb func()) {
	ch.thumbClickCallbacks = append(ch.thumbClickCallbacks, cb)
}

func (ch *ChartThumbnail) Close() {
	ch.lines.Close()
	ch.lines = nil
	ch.dailyStochastics.Close()
	ch.dailyStochastics = nil
	ch.weeklyStochastics.Close()
	ch.weeklyStochastics = nil
}
