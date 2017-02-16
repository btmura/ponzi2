package ponzi

import (
	"fmt"
	"time"
)

// newYorkLoc is the New York timezone.
var newYorkLoc = mustLoadLocation("America/New_York")

// mustLoadLocation loads the requested tz location or panics.
func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(fmt.Sprintf("time.LoadLocation: %v", err))
	}
	return loc
}

// midnight returns the given time with the hours, minutes, seconds, and nanoseconds set to zero.
func midnight(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// sortableTimes is a time.Time slice that can be sorted.
type sortableTimes []time.Time

// Len implements sort.Interface
func (st sortableTimes) Len() int {
	return len(st)
}

// Less implements sort.Interface
func (st sortableTimes) Less(i, j int) bool {
	return st[i].Before(st[j])
}

// Swap implements sort.Interface
func (st sortableTimes) Swap(i, j int) {
	st[i], st[j] = st[j], st[i]
}
