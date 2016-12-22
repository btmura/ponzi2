package ponzi

// view describes how to render the model to the screen.
type view struct {
	model *model
}

func (v *view) dowPriceText() string {
	return "123"
}

func (v *view) sapPriceText() string {
	return "456"
}

func (v *view) nasdaqPriceText() string {
	return "789"
}
