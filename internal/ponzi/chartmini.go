package ponzi

import (
	"image"
)

type miniChart struct {
	stock             *modelStock
	frameBorder       *vao
	frameDivider      *vao
	dailyStochastics  *chartStochastics
	weeklyStochastics *chartStochastics
}

func createMiniChart(stock *modelStock, labelText *dynamicText) *miniChart {
	return &miniChart{
		stock:             stock,
		frameBorder:       createStrokedRectVAO(white, white, white, white),
		frameDivider:      createLineVAO(white, white),
		dailyStochastics:  createChartStochastics(stock, labelText, daily),
		weeklyStochastics: createChartStochastics(stock, labelText, weekly),
	}
}

func (mc *miniChart) update() {
	mc.dailyStochastics.update()
	mc.weeklyStochastics.update()
}

func (mc *miniChart) render(r image.Rectangle) {
	setModelMatrixRectangle(r)
	mc.frameBorder.render()

	const pad = 5
	subRects := sliceRectangle(r, 0.5, 0.5)
	mc.dailyStochastics.render(subRects[1].Inset(pad))
	mc.weeklyStochastics.render(subRects[0].Inset(pad))

	setModelMatrixRectangle(image.Rect(r.Min.X, r.Min.Y+r.Dy()/2, r.Max.X, r.Min.Y+r.Dy()/2))
	mc.frameDivider.render()
}

func (mc *miniChart) close() {
	mc.frameBorder.close()
	mc.frameBorder = nil
	mc.frameDivider.close()
	mc.frameDivider = nil
	mc.dailyStochastics.close()
	mc.dailyStochastics = nil
	mc.weeklyStochastics.close()
	mc.weeklyStochastics = nil
}
