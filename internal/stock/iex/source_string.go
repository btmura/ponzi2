// Code generated by "stringer -type=Source"; DO NOT EDIT.

package iex

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[SourceUnspecified-0]
	_ = x[RealTimePrice-1]
	_ = x[FifteenMinuteDelayedPrice-2]
	_ = x[Close-3]
	_ = x[PreviousClose-4]
	_ = x[Price-5]
	_ = x[LastTrade-6]
}

const _Source_name = "SourceUnspecifiedRealTimePriceFifteenMinuteDelayedPriceClosePreviousClosePriceLastTrade"

var _Source_index = [...]uint8{0, 17, 30, 55, 60, 73, 78, 87}

func (i Source) String() string {
	if i < 0 || i >= Source(len(_Source_index)-1) {
		return "Source(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Source_name[_Source_index[i]:_Source_index[i+1]]
}
