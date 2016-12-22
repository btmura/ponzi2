package ponzi

import "fmt"

// view describes how to render the model to the screen.
type view struct {
	model *model
}

func (v *view) dowPriceText() string {
	return formatQuote(v.model.dow)
}

func (v *view) sapPriceText() string {
	return formatQuote(v.model.sap)
}

func (v *view) nasdaqPriceText() string {
	return formatQuote(v.model.nasdaq)
}

func formatQuote(q *quote) string {
	if q != nil {
		return fmt.Sprintf("%10.2f %+5.2f %+5.2f%%", q.price, q.change, q.percentChange*100.0)
	}
	return "..."
}
