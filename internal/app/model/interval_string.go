// Code generated by "stringer -type=Interval"; DO NOT EDIT.

package model

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[IntervalUnspecified-0]
	_ = x[Intraday-1]
	_ = x[Daily-2]
}

const _Interval_name = "IntervalUnspecifiedIntradayDaily"

var _Interval_index = [...]uint8{0, 19, 27, 32}

func (i Interval) String() string {
	if i < 0 || i >= Interval(len(_Interval_index)-1) {
		return "Interval(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Interval_name[_Interval_index[i]:_Interval_index[i+1]]
}
