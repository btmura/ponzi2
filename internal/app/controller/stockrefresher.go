package controller

import (
	"context"
	"time"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/errors"
	"github.com/btmura/ponzi2/internal/log"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

type stockRefresher struct {
	// iexClient fetches stock data to update the model.
	iexClient iexClientInterface

	// eventController allows the stockRefresher to post stock updates.
	eventController *eventController

	// refreshTicker ticks to trigger refreshes during market hours.
	refreshTicker *time.Ticker

	// enabled enables refreshing stocks when set to true.
	enabled bool
}

func newStockRefresher(iexClient iexClientInterface, eventController *eventController) *stockRefresher {
	return &stockRefresher{
		iexClient:       iexClient,
		eventController: eventController,
		refreshTicker:   time.NewTicker(5 * time.Minute),
	}
}

// refreshLoop refreshes stocks during market hours.
func (s *stockRefresher) refreshLoop() {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatalf("time.LoadLocation: %v", err)
	}

	for t := range s.refreshTicker.C {
		n := time.Now()
		open := time.Date(n.Year(), n.Month(), n.Day(), 9, 30, 0, 0, loc)
		close := time.Date(n.Year(), n.Month(), n.Day(), 16, 0, 0, 0, loc)

		if t.Before(open) || t.After(close) {
			log.Infof("ignoring refresh ticker at %v", t.Format("1/2/2006 3:04:05 PM"))
			continue
		}

		s.eventController.addEventLocked(event{refreshAllStocks: true})
	}
}
func (s *stockRefresher) start() {
	s.enabled = true
}

func (s *stockRefresher) stop() {
	s.enabled = false
	s.refreshTicker.Stop()
}

func (s *stockRefresher) refreshOne(ctx context.Context, symbol string, dataRange model.Range) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	if dataRange == model.RangeUnspecified {
		return errors.Errorf("range not set")
	}

	d := new(dataRequestBuilder)
	if err := d.add([]string{symbol}, dataRange); err != nil {
		return err
	}
	return s.refresh(ctx, d)
}

func (s *stockRefresher) refresh(ctx context.Context, d *dataRequestBuilder) error {
	if !s.enabled {
		log.Infof("ignoring stock refresh request, refreshing disabled")
		return nil
	}

	reqs, err := d.dataRequests()
	if err != nil {
		return err
	}

	for _, req := range reqs {
		for _, sym := range req.symbols {
			s.eventController.addEventLocked(event{
				symbol:         sym,
				dataRange:      req.dataRange,
				refreshStarted: true,
			})
		}
	}

	for _, req := range reqs {
		go func(req *dataRequest) {
			handleErr := func(err error) {
				var es []event
				for _, sym := range req.symbols {
					es = append(es, event{
						symbol:    sym,
						updateErr: err,
					})
				}
				s.eventController.addEventLocked(es...)
			}

			quotes, err := s.iexClient.GetQuotes(ctx, req.quotesRequest)
			if err != nil {
				handleErr(err)
				return
			}

			charts, err := s.iexClient.GetCharts(ctx, req.chartsRequest)
			if err != nil {
				handleErr(err)
				return
			}

			type stockData struct {
				quote *iex.Quote
				chart *iex.Chart
			}

			symbol2StockData := map[string]*stockData{}

			for _, q := range quotes {
				d := symbol2StockData[q.Symbol]
				if d == nil {
					d = &stockData{}
					symbol2StockData[q.Symbol] = d
				}
				d.quote = q
			}

			for _, ch := range charts {
				d := symbol2StockData[ch.Symbol]
				if d == nil {
					d = &stockData{}
					symbol2StockData[ch.Symbol] = d
				}
				d.chart = ch
			}

			var es []event

			for sym, stockData := range symbol2StockData {
				q, err := modelQuote(stockData.quote)
				if err != nil {
					es = append(es, event{
						symbol:    sym,
						updateErr: err,
					})
					continue
				}

				switch req.dataRange {
				case model.OneDay:
					ch, err := modelOneDayChart(stockData.chart)
					es = append(es, event{
						symbol:    sym,
						quote:     q,
						chart:     ch,
						updateErr: err,
					})

				case model.OneYear:
					ch, err := modelOneYearChart(stockData.quote, stockData.chart)
					es = append(es, event{
						symbol:    sym,
						quote:     q,
						chart:     ch,
						updateErr: err,
					})
				}
			}

			for _, sym := range req.symbols {
				if symbol2StockData[sym] != nil {
					continue
				}
				es = append(es, event{
					symbol:    sym,
					updateErr: errors.Errorf("no stock data for %q", sym),
				})
			}

			s.eventController.addEventLocked(es...)
		}(req)
	}

	return nil
}

// dataRequestBuilder accumulates symbols and data ranges and builds a slice of data requests.
type dataRequestBuilder struct {
	range2Symbols map[model.Range][]string
}

func (d *dataRequestBuilder) add(symbols []string, dataRange model.Range) error {
	for _, s := range symbols {
		if err := model.ValidateSymbol(s); err != nil {
			return err
		}
	}

	if dataRange == model.RangeUnspecified {
		return errors.Errorf("range not set")
	}

	if len(symbols) == 0 {
		return nil
	}

	sset := make(map[string]bool)
	for _, s := range d.range2Symbols[dataRange] {
		sset[s] = true
	}
	for _, s := range symbols {
		sset[s] = true
	}

	var ss []string
	for s := range sset {
		ss = append(ss, s)
	}

	if d.range2Symbols == nil {
		d.range2Symbols = make(map[model.Range][]string)
	}
	d.range2Symbols[dataRange] = ss

	return nil
}

type dataRequest struct {
	symbols       []string
	dataRange     model.Range
	quotesRequest *iex.GetQuotesRequest
	chartsRequest *iex.GetChartsRequest
}

func (d *dataRequestBuilder) dataRequests() ([]*dataRequest, error) {
	var reqs []*dataRequest
	for r, ss := range d.range2Symbols {
		var ir iex.Range

		switch r {
		case model.OneDay:
			ir = iex.OneDay
		case model.OneYear:
			ir = iex.TwoYears // Need additional data for weekly stochastics.
		default:
			return nil, errors.Errorf("bad range: %v", r)
		}

		reqs = append(reqs, &dataRequest{
			symbols:   ss,
			dataRange: r,
			quotesRequest: &iex.GetQuotesRequest{
				Symbols: ss,
			},
			chartsRequest: &iex.GetChartsRequest{
				Symbols: ss,
				Range:   ir,
			},
		})
	}
	return reqs, nil
}
