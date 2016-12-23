package ponzi

import (
	"fmt"
	"log"

	"github.com/go-gl/glfw/v3.2/glfw"
)

// view describes how to render the model to the screen.
type view struct {
	model *model
}

func (v *view) handleKey(key glfw.Key, action glfw.Action) {
	log.Printf("key: %v", key)
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
		return fmt.Sprintf(" %.2f %+5.2f %+5.2f%% ", q.price, q.change, q.percentChange*100.0)
	}
	return "..."
}
