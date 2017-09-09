// Package time has time utility functions.
package time

import (
	"flag"
	"time"

	"github.com/golang/glog"
)

func init() {
	flag.Parse() // Avoid glog errors about logging before flag.Parse.
}

// NewYorkLoc is the New York timezone.
var NewYorkLoc = mustLoadLocation("America/New_York")

// mustLoadLocation loads the requested tz location or panics.
func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		glog.Fatalf("mustLoadLocation: failed to load location: %v", err)
	}
	return loc
}

// Midnight returns the given time with the hours, minutes, seconds, and nanoseconds set to zero.
func Midnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
