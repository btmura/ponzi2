package alpha

import (
	"bytes"
	"context"
	"fmt"
	"strings"
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
	fpath := fmt.Sprintf("/data/demo-movavg-%d.txt", req.TimePeriod)
	return decodeMovingAverageResponse(bytes.NewReader(_escFSMustByte(false, fpath)))
}

// GetStochastics returns Stochastics or an error.
func (d *Demo) GetStochastics(ctx context.Context, req *GetStochasticsRequest) (*Stochastics, error) {
	fpath := fmt.Sprintf("/data/demo-stoch-%s.txt", strings.ToLower(req.Interval.String()))
	return decodeStochasticsResponse(bytes.NewReader(_escFSMustByte(false, fpath)))
}
