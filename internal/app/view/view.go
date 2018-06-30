package view

import (
	"image"
	"log"
	"os"

	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/gfx"
)

// Get esc from github.com/mjibson/esc. It's used to embed resources into the binary.
//go:generate esc -o bindata.go -pkg view -include ".*(ply|png)" -modtime 1337 -private data

var logger = log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile)

// Frames per second.
// TODO(btmura): remove duplicate fps constant in controller
const fps = 60.0

// Colors used throughout the UI.
var (
	green     = [3]float32{0.25, 1, 0}
	red       = [3]float32{1, 0.3, 0}
	yellow    = [3]float32{1, 1, 0}
	purple    = [3]float32{0.5, 0, 1}
	white     = [3]float32{1, 1, 1}
	gray      = [3]float32{0.15, 0.15, 0.15}
	lightGray = [3]float32{0.35, 0.35, 0.35}
	orange    = [3]float32{1, 0.5, 0}
)

const viewOuterPadding = 10

var viewChartThumbSize = image.Pt(155, 105)

var (
	viewInputSymbolTextRenderer = gfx.NewTextRenderer(goregular.TTF, 48)
	viewInstructionsText        = NewCenteredText(gfx.NewTextRenderer(goregular.TTF, 24), "Type in symbol and press ENTER...")
)

// The View renders the UI to view and edit the model's stocks that it observes.
type View struct {
	// chart renders the currently viewed stock.
	chart *Chart

	// chartThumbs renders the stocks in the sidebar.
	chartThumbs []*ChartThumb

	// inputSymbol stores and renders the symbol being entered by the user.
	inputSymbol *CenteredText
}

// ViewContext is passed down the view hierarchy providing drawing hints and
// event information. Meant to be passed around like a Rectangle or Point rather
// than a pointer to avoid mistakes.
type ViewContext struct {
	// Bounds is the rectangle with global coords that should be drawn within.
	Bounds image.Rectangle

	// MousePos is the current global mouse position.
	MousePos image.Point

	// MouseLeftButtonClicked is whether the left mouse button was clicked.
	MouseLeftButtonClicked bool

	// Fudge is the position from 0 to 1 between the current and next frame.
	Fudge float32

	// ScheduledCallbacks are callbacks to be called at the end of Render.
	ScheduledCallbacks *[]func()
}

// LeftClickInBounds returns true if the left mouse button was clicked within
// the context's bounds. Doesn't take into account overlapping view parts.
func (vc ViewContext) LeftClickInBounds() bool {
	return vc.MouseLeftButtonClicked && vc.MousePos.In(vc.Bounds)
}

// NewView creates a new View.
func NewView() *View {
	return &View{
		inputSymbol: NewCenteredText(viewInputSymbolTextRenderer, "", CenteredTextBubble(chartRounding, chartPadding)),
	}
}

// Update updates the View.
func (v *View) Update() {
	if v.chart != nil {
		v.chart.Update()
	}
	for _, th := range v.chartThumbs {
		th.Update()
	}
}

// Render renders the View.
func (v *View) Render(vc ViewContext) {
	ogBnds := vc.Bounds.Inset(viewOuterPadding)

	// Calculate bounds for main area.
	vc.Bounds = ogBnds
	if len(v.chartThumbs) > 0 {
		vc.Bounds.Min.X += viewOuterPadding + viewChartThumbSize.X
	}

	// Render the the main chart or instructions.
	if v.chart != nil {
		v.chart.Render(vc)
	} else {
		viewInstructionsText.Render(vc.Bounds)
	}

	// Render the input symbol over the chart.
	v.inputSymbol.Render(vc.Bounds)

	// Render the sidebar thumbnails.
	vc.Bounds = image.Rect(
		viewOuterPadding, ogBnds.Max.Y-viewChartThumbSize.Y,
		viewOuterPadding+viewChartThumbSize.X, ogBnds.Max.Y,
	)
	for _, th := range v.chartThumbs {
		th.Render(vc)
		vc.Bounds = vc.Bounds.Sub(image.Pt(0, viewChartThumbSize.Y+viewOuterPadding))
	}
}

// SetChart sets the View's main chart.
func (v *View) SetChart(ch *Chart) {
	v.chart = ch
}

// AddChartThumb adds the ChartThumbnail to the side bar.
func (v *View) AddChartThumb(th *ChartThumb) {
	v.chartThumbs = append(v.chartThumbs, th)
}

// RemoveChartThumb removes the ChartThumbnail from the side bar.
func (v *View) RemoveChartThumb(th *ChartThumb) {
	for i, thumb := range v.chartThumbs {
		if thumb == th {
			v.chartThumbs = append(v.chartThumbs[:i], v.chartThumbs[i+1:]...)
			break
		}
	}
}

// InputSymbol returns the input symbol entered by the user.
func (v *View) InputSymbol() string {
	return v.inputSymbol.Text
}

// ClearInputSymbol clears the input symbol entered by the user.
func (v *View) ClearInputSymbol() {
	v.inputSymbol.Text = ""
}

// PushInputSymbolChar pushes the character to the input symbol.
func (v *View) PushInputSymbolChar(ch rune) {
	v.inputSymbol.Text += string(ch)
}

// PopInputSymbolChar pops off the last character of the input symbol.
func (v *View) PopInputSymbolChar() {
	if l := len(v.inputSymbol.Text); l > 0 {
		v.inputSymbol.Text = v.inputSymbol.Text[:l-1]
	}
}
