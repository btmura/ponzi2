package ponzi

import (
	"image"
	"unicode"

	"github.com/go-gl/gl/v4.5-core/gl"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/gfx"
	math2 "github.com/btmura/ponzi2/internal/math"
)

// Colors used throughout the UI.
var (
	green  = [3]float32{0.25, 1, 0}
	red    = [3]float32{1, 0.3, 0}
	yellow = [3]float32{1, 1, 0}
	purple = [3]float32{0.5, 0, 1}
	white  = [3]float32{1, 1, 1}
	gray   = [3]float32{0.15, 0.15, 0.15}
)

var inputSymbolTextRenderer = gfx.NewTextRenderer(goregular.TTF, 48)

const viewOuterPadding = 10

var viewChartThumbSize = image.Pt(140, 90)

// The View renders the UI to view and edit the model's stocks that it observes.
type View struct {
	// chart renders the currently viewed stock.
	chart *Chart

	// chartThumbs renders the stocks in the sidebar.
	chartThumbs []*ChartThumbnail

	// inputSymbol stores and renders the symbol being entered by the user.
	inputSymbol *CenteredText
}

// ViewContext is passed down the view hierarchy providing drawing hints and event information.
// Meant to be passed around like a Rectangle or Point rather than a pointer to avoid mistakes.
type ViewContext struct {
	// Bounds is the rectangle with global coordinates that the view part should draw within.
	Bounds image.Rectangle

	// MousePos is the current global mouse position.
	MousePos image.Point

	// MouseLeftButtonClicked is whether the left mouse button was clicked.
	MouseLeftButtonClicked bool

	// Fudge is the position from 0 to 1 between the current and next animation frame.
	Fudge float32

	// values stores values collected throughout the Render pass.
	values *viewContextValues
}

// viewContextValues stores values collected throughout the Render pass.
type viewContextValues struct {
	// scheduledCallbacks are callbacks that should be called at the end of Render.
	scheduledCallbacks []func()
}

// LeftClickInBounds returns true if the left mouse button was clicked within the context's bounds.
// Doesn't take into account overlapping view parts.
func (vc ViewContext) LeftClickInBounds() bool {
	return vc.MouseLeftButtonClicked && vc.MousePos.In(vc.Bounds)
}

// ScheduleCallback schedules a callback that will be called after Render is done.
func (vc ViewContext) ScheduleCallback(cb func()) {
	vc.values.scheduledCallbacks = append(vc.values.scheduledCallbacks, cb)
}

// NewView creates a new View.
func NewView() *View {
	return &View{
		inputSymbol: NewCenteredText(inputSymbolTextRenderer, ""),
	}
}

// Render renders the view.
func (v *View) Render(vc ViewContext) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// Render the input symbol and the main chart.
	if v.chart != nil {
		vc.Bounds = vc.Bounds.Inset(viewOuterPadding)
		if len(v.chartThumbs) > 0 {
			vc.Bounds.Min.X += viewOuterPadding + viewChartThumbSize.X
		}
		v.inputSymbol.Render(vc)
		v.chart.Render(vc)
	}

	// Render the sidebar thumbnails.
	vc.Bounds = image.Rect(
		viewOuterPadding, vc.Bounds.Max.Y-viewChartThumbSize.Y,
		viewOuterPadding+viewChartThumbSize.X, vc.Bounds.Max.Y,
	)
	for _, th := range v.chartThumbs {
		th.Render(vc)
		vc.Bounds = vc.Bounds.Sub(image.Pt(0, viewChartThumbSize.Y+viewOuterPadding))
	}
}

func (v *View) SetChart(st *ModelStock) *Chart {
	ch := NewChart()
	ch.Update(st)
	st.AddChangeCallback(func() {
		ch.Update(st)
	})
	v.chart = ch
	return ch
}

func (v *View) AddChartThumb(st *ModelStock) *ChartThumbnail {
	th := NewChartThumbnail()
	th.Update(st)
	st.AddChangeCallback(func() {
		th.Update(st)
	})
	v.chartThumbs = append(v.chartThumbs, th)
	return th
}

func (v *View) RemoveChartThumb(th *ChartThumbnail) {
	for i, thumb := range v.chartThumbs {
		if thumb == th {
			v.chartThumbs = append(v.chartThumbs[:i], v.chartThumbs[i+1:]...)
			break
		}
	}
	th.Close()
}

// Resize responds to window size changes by updating internal matrices.
func (v *View) Resize(newSize image.Point) {
	gl.Viewport(0, 0, int32(newSize.X), int32(newSize.Y))

	// Calculate the new ortho projection view matrix.
	fw, fh := float32(newSize.X), float32(newSize.Y)
	gfx.SetProjectionViewMatrix(math2.OrthoMatrix(fw, fh, fw /* use width as depth */))
}

func (v *View) InputSymbol() string {
	return v.inputSymbol.Text
}

func (v *View) ClearInputSymbol() {
	v.inputSymbol.Text = ""
}

func (v *View) PushInputSymbolChar(ch rune) (pushed bool) {
	ch = unicode.ToUpper(ch)
	if _, ok := acceptedChars[ch]; ok {
		v.inputSymbol.Text += string(ch)
		return true
	}
	return false
}

func (v *View) PopInputSymbolChar() {
	if l := len(v.inputSymbol.Text); l > 0 {
		v.inputSymbol.Text = v.inputSymbol.Text[:l-1]
	}
}
