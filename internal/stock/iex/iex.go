// Package iex provides a client to get stock data using the IEX API.
package iex

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"time"
)

// Internal package variables used for the implementation.
var (
	// now is a function to get the current time. Mocked out in tests to return a fixed time.
	now = time.Now

	// loc is the timezone to use when parsing dates.
	loc = mustLoadLocation("America/New_York")
)

// Range is the range to specify in the request.
type Range int

// Range values.
//go:generate stringer -type=Range
const (
	RangeUnspecified Range = iota
	OneDay
	TwoYears
)

// Client is used to make IEX API requests.
type Client struct {
	// token is the API token required on all requests.
	token string

	// dumpAPIResponses dumps API responses into text files.
	dumpAPIResponses bool
}

// NewClient returns a new Client.
func NewClient(token string, dumpAPIResponses bool) *Client {
	return &Client{token: token, dumpAPIResponses: dumpAPIResponses}
}

func millisToTime(ms int64) time.Time {
	sec := ms / 1e3
	nsec := ms*1e6 - sec*1e9
	return time.Unix(sec, nsec)
}

func dumpResponse(fileName string, r io.Reader) (io.ReadCloser, error) {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0660)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(file, "%s", b)

	return ioutil.NopCloser(bytes.NewBuffer(b)), nil
}

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		log.Fatalf("time.LoadLocation(%s) failed: %v", name, err)
	}
	return loc
}
