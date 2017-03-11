package ponzi

import "fmt"

func (ch *chart) maxStochasticLabelWidth() int {
	width := func(percent float32) int {
		s := ch.labelText.measure(stochasticLabelText(percent))
		return s.X + chartLabelPadding*2
	}

	w1, w2 := width(.7), width(.3)
	if w1 > w2 {
		return w1
	}
	return w2
}

func stochasticLabelText(percent float32) string {
	return fmt.Sprintf("%.f%%", percent*100)
}
