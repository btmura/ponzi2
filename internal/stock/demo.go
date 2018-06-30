package stock

import (
	"bytes"
	"context"
	"fmt"
)

// Demo returns predefined offline data.
type Demo struct{}

// NewDemo returns a new Demo.
func NewDemo() *Demo {
	return &Demo{}
}

// GetHistory returns stock data or an error.
func (d *Demo) GetHistory(ctx context.Context, req *GetHistoryRequest) (*History, error) {
	return decodeHistoryResponse(bytes.NewReader(_escFSMustByte(false, "/data/demo-hist.txt")))
}

// GetMovingAverage returns MovingAverage data or an error.
func (d *Demo) GetMovingAverage(ctx context.Context, req *GetMovingAverageRequest) (*MovingAverage, error) {
	var fpath string
	switch req.TimePeriod {
	case 25:
		fpath = "/data/demo-movavg-25.txt"
	case 50:
		fpath = "/data/demo-movavg-50.txt"
	case 250:
		fpath = "/data/demo-movang-250.txt"
	default:
		return nil, fmt.Errorf("stock: unsupported demo movavg time period: %d", req.TimePeriod)
	}
	return decodeMovingAverageResponse(bytes.NewReader(_escFSMustByte(false, fpath)))
}

// GetStochastics returns Stochastics or an error.
func (d *Demo) GetStochastics(ctx context.Context, req *GetStochasticsRequest) (*Stochastics, error) {
	var fpath string
	switch req.Interval {
	case Daily:
		fpath = "/data/demo-stoch-daily.txt"
	case Weekly:
		fpath = "/data/demo-stoch-weekly.txt"
	default:
		return nil, fmt.Errorf("stock: unsupported demo stoch interval: %d", req.Interval)
	}
	return decodeStochasticsResponse(bytes.NewReader(_escFSMustByte(false, fpath)))
}
