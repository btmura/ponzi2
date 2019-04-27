package controller

import (
	"context"
	"log"
	"time"

	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/errors"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

type stockRefresher struct {
	// iexClient fetches stock data to update the model.
	iexClient *iex.Client

	// eventController allows the stockRefresher to post stock updates.
	eventController *eventController

	// refreshTicker ticks to trigger refreshes during market hours.
	refreshTicker *time.Ticker

	// enabled enables refreshing stocks when set to true.
	enabled bool
}

func newStockRefresher(iexClient *iex.Client, eventController *eventController) *stockRefresher {
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
			glog.V(2).Infof("ignoring refresh ticker at %v", t.Format("1/2/2006 3:04:05 PM"))
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
		glog.V(2).Infof("ignoring stock refresh request, refreshing disabled")
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

			stocks, err := s.iexClient.GetStocks(ctx, req.iexRequest)
			if err != nil {
				handleErr(err)
				return
			}

			var es []event

			found := map[string]bool{}
			for _, st := range stocks {
				found[st.Symbol] = true

				switch req.dataRange {
				case model.OneDay:
					ch, err := modelOneDayChart(st)
					es = append(es, event{
						symbol:    st.Symbol,
						chart:     ch,
						updateErr: err,
					})

				case model.OneYear:
					ch, err := modelOneYearChart(st)
					es = append(es, event{
						symbol:    st.Symbol,
						chart:     ch,
						updateErr: err,
					})
				}
			}

			for _, sym := range req.symbols {
				if found[sym] {
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
	symbols    []string
	dataRange  model.Range
	iexRequest *iex.GetStocksRequest
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
			iexRequest: &iex.GetStocksRequest{
				Symbols: ss,
				Range:   ir,
			},
		})
	}
	return reqs, nil
}
