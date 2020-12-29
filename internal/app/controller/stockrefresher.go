package controller

import (
	"context"
	"time"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/errs"
	"github.com/btmura/ponzi2/internal/logger"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

type stockRefresher struct {
	// iexClient fetches stock data to update the model.
	iexClient iexClientInterface

	// token is the IEX API token to be included on requests.
	token string

	// eventController allows the stockRefresher to post stock updates.
	eventController *eventController

	// refreshTicker ticks to trigger refreshes during market hours.
	refreshTicker *time.Ticker

	// enabled enables refreshing stocks when set to true.
	enabled bool
}

func newStockRefresher(iexClient iexClientInterface, token string, eventController *eventController) *stockRefresher {
	return &stockRefresher{
		iexClient:       iexClient,
		token:           token,
		eventController: eventController,
		refreshTicker:   time.NewTicker(5 * time.Minute),
	}
}

// refreshLoop refreshes stocks during market hours.
func (s *stockRefresher) refreshLoop() {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		logger.Fatalf("time.LoadLocation: %v", err)
	}

	for t := range s.refreshTicker.C {
		n := time.Now()
		openTime := time.Date(n.Year(), n.Month(), n.Day(), 9, 30, 0, 0, loc)
		closeTime := time.Date(n.Year(), n.Month(), n.Day(), 16, 0, 0, 0, loc)

		if openTime.Weekday() == time.Saturday || openTime.Weekday() == time.Sunday || t.Before(openTime) || t.After(closeTime) {
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

func (s *stockRefresher) refreshOne(ctx context.Context, symbol string, interval model.Interval) error {
	if err := model.ValidateSymbol(symbol); err != nil {
		return err
	}

	if interval == model.IntervalUnspecified {
		return errs.Errorf("unspecified interval")
	}

	d := new(dataRequestBuilder)
	if err := d.add([]string{symbol}, interval); err != nil {
		return err
	}
	return s.refresh(ctx, d)
}

func (s *stockRefresher) refresh(ctx context.Context, d *dataRequestBuilder) error {
	if !s.enabled {
		return nil
	}

	reqs, err := d.dataRequests(s.token)
	if err != nil {
		return err
	}

	for _, req := range reqs {
		for _, sym := range req.symbols {
			for _, interval := range req.intervals {
				s.eventController.addEventLocked(event{
					symbol:         sym,
					interval:       interval,
					refreshStarted: true,
				})
			}
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

				for _, interval := range req.intervals {
					switch interval {
					case model.Intraday:
						ch := modelIntradayChart(stockData.chart)
						es = append(es, event{
							symbol: sym,
							quote:  q,
							chart:  ch,
						})

					case model.Daily:
						ch := modelDailyChart(stockData.quote, stockData.chart)
						es = append(es, event{
							symbol: sym,
							quote:  q,
							chart:  ch,
						})

					case model.Weekly:
						ch := modelWeeklyChart(stockData.quote, stockData.chart)
						es = append(es, event{
							symbol: sym,
							quote:  q,
							chart:  ch,
						})
					}
				}
			}

			for _, sym := range req.symbols {
				if symbol2StockData[sym] != nil {
					continue
				}
				es = append(es, event{
					symbol:    sym,
					updateErr: errs.Errorf("no stock data for %q", sym),
				})
			}

			s.eventController.addEventLocked(es...)
		}(req)
	}

	return nil
}

// dataRequestBuilder accumulates symbols into request groups and then builds the requests.
type dataRequestBuilder struct {
	symbolGroups map[dataRequestGroup][]string
}

type dataRequestGroup int

const (
	dataRequestGroupUnspecified dataRequestGroup = iota
	intraday
	dailyWeekly
)

func (d dataRequestGroup) Intervals() []model.Interval {
	switch d {
	case dataRequestGroupUnspecified:
		return nil
	case intraday:
		return []model.Interval{model.Intraday}
	case dailyWeekly:
		return []model.Interval{model.Daily, model.Weekly}
	default:
		return nil
	}
}

var interval2DataRequestGroup = map[model.Interval]dataRequestGroup{
	model.Intraday: intraday,
	model.Daily:    dailyWeekly,
	model.Weekly:   dailyWeekly,
}

func (d *dataRequestBuilder) add(symbols []string, interval model.Interval) error {
	if len(symbols) == 0 {
		return nil
	}

	for _, s := range symbols {
		if err := model.ValidateSymbol(s); err != nil {
			return err
		}
	}

	if interval == model.IntervalUnspecified {
		return errs.Errorf("unspecified interval")
	}

	group := interval2DataRequestGroup[interval]

	symbolSet := make(map[string]bool)
	for _, s := range d.symbolGroups[group] {
		symbolSet[s] = true
	}
	for _, s := range symbols {
		symbolSet[s] = true
	}

	var ss []string
	for s := range symbolSet {
		ss = append(ss, s)
	}

	if d.symbolGroups == nil {
		d.symbolGroups = make(map[dataRequestGroup][]string)
	}
	d.symbolGroups[group] = ss

	return nil
}

type dataRequest struct {
	symbols       []string
	intervals     []model.Interval
	quotesRequest *iex.GetQuotesRequest
	chartsRequest *iex.GetChartsRequest
}

func (d *dataRequestBuilder) dataRequests(token string) ([]*dataRequest, error) {
	var reqs []*dataRequest
	for group, ss := range d.symbolGroups {
		var iexRange iex.Range

		switch group {
		case dailyWeekly:
			iexRange = iex.TwoYears
		default:
			return nil, errs.Errorf("bad group: %v", group)
		}

		reqs = append(reqs, &dataRequest{
			symbols:   ss,
			intervals: group.Intervals(),
			quotesRequest: &iex.GetQuotesRequest{
				Token:   token,
				Symbols: ss,
			},
			chartsRequest: &iex.GetChartsRequest{
				Token:   token,
				Symbols: ss,
				Range:   iexRange,
			},
		})
	}
	return reqs, nil
}
