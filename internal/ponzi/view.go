package ponzi

import (
	"image"
	"time"
	"unicode"

	"github.com/go-gl/gl/v4.5-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang/glog"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/btmura/ponzi2/internal/gfx"
	math2 "github.com/btmura/ponzi2/internal/math"
	"github.com/btmura/ponzi2/internal/stock"
	time2 "github.com/btmura/ponzi2/internal/time"
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

// acceptedChars are the chars the user can enter for a symbol.
var acceptedChars = map[rune]bool{
	'A': true, 'B': true, 'C': true,
	'D': true, 'E': true, 'F': true,
	'G': true, 'H': true, 'I': true,
	'J': true, 'K': true, 'L': true,
	'M': true, 'N': true, 'O': true,
	'P': true, 'Q': true, 'R': true,
	'S': true, 'T': true, 'U': true,
	'V': true, 'W': true, 'X': true,
	'Y': true, 'Z': true,
}

var inputSymbolTextRenderer = gfx.NewTextRenderer(goregular.TTF, 48)

const viewOuterPadding = 10

var viewChartThumbSize = image.Pt(140, 90)

// The View renders the UI to view and edit the model's stocks that it observes.
type View struct {
	// model has the stock data to be rendered.
	model *Model

	// chart renders the model's current stock.
	chart *Chart

	// chartThumbs renders the model's other stocks.
	chartThumbs []*ChartThumbnail

	// inputSymbol stores and renders the symbol being entered by the user.
	inputSymbol *CenteredText

	// pendingStockUpdates is a channel with stock updates ready to apply to the model.
	pendingStockUpdates chan viewStockUpdate

	// mousePos is the current global mouse position.
	mousePos image.Point

	// mouseLeftButtonClicked is whether the left mouse button was clicked.
	mouseLeftButtonClicked bool

	// winSize is the current window's size used to measure and draw the UI.
	winSize image.Point
}

// viewStockUpdate bundles a stock and new data for that stock.
type viewStockUpdate struct {
	// stock is the stock to update.
	stock *ModelStock

	// tradingHistory is the new data to update the stock with.
	tradingHistory *stock.TradingHistory
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

// NewView creates a new View that observes the given Model.
func NewView(model *Model) *View {
	if err := gl.Init(); err != nil {
		glog.Fatalf("newView: failed to init OpenGL: %v", err)
	}
	glog.Infof("OpenGL version: %s", gl.GoStr(gl.GetString(gl.VERSION)))

	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(0, 0, 0, 0)

	if err := gfx.InitProgram(); err != nil {
		glog.Fatalf("newView: failed to init gfx: %v", err)
	}

	return &View{
		model:               model,
		inputSymbol:         NewCenteredText(inputSymbolTextRenderer, ""),
		pendingStockUpdates: make(chan viewStockUpdate),
	}
}

// Render renders the view.
func (v *View) Render(fudge float32) {
	// Process any stock updates.
loop:
	for {
		select {
		case u := <-v.pendingStockUpdates:
			u.stock.Update(u.tradingHistory)

		default:
			break loop
		}
	}

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	vc := ViewContext{
		MousePos:               v.mousePos,
		MouseLeftButtonClicked: v.mouseLeftButtonClicked,
		values:                 &viewContextValues{},
	}

	// Render the input symbol and the main chart.
	if v.chart != nil {
		vc.Bounds = image.Rectangle{image.ZP, v.winSize}.Inset(viewOuterPadding)
		if len(v.model.SavedStocks) > 0 {
			vc.Bounds.Min.X += viewOuterPadding + viewChartThumbSize.X
		}
		v.inputSymbol.Render(vc)
		v.chart.Render(vc)
	}

	// Render the sidebar thumbnails.
	for i, th := range v.chartThumbs {
		min := image.Pt(viewOuterPadding, 0)
		max := image.Pt(min.X+viewChartThumbSize.X, 0)
		max.Y = v.winSize.Y - (viewOuterPadding+viewChartThumbSize.Y)*i - viewOuterPadding
		min.Y = max.Y - viewChartThumbSize.Y
		vc.Bounds = image.Rectangle{min, max}
		th.Render(vc)
	}

	// Call any callbacks scheduled by views.
	for _, cb := range vc.values.scheduledCallbacks {
		cb()
	}

	// Reset any flags for the next viewContext.
	v.mouseLeftButtonClicked = false
}

func (v *View) SetChart(st *ModelStock) {
	ch := NewChart()
	ch.Update(st)
	st.AddChangeCallback(func() {
		ch.Update(st)
	})
	v.chart = ch

	ch.SetAddButtonClickCallback(func() {
		st, added := v.model.AddSavedStock(st.Symbol)
		if !added {
			return
		}
		v.AddChartThumb(st)
		v.goSaveConfig()
	})
}

func (v *View) AddChartThumb(st *ModelStock) {
	th := NewChartThumbnail()
	th.Update(st)
	st.AddChangeCallback(func() {
		th.Update(st)
	})
	v.chartThumbs = append(v.chartThumbs, th)

	th.SetRemoveButtonClickCallback(func() {
		removed := v.model.RemoveSavedStock(st.Symbol)
		if !removed {
			return
		}

		for i, thumb := range v.chartThumbs {
			if thumb == th {
				v.chartThumbs = append(v.chartThumbs[:i], v.chartThumbs[i+1:]...)
				break
			}
		}
		th.Close()

		v.goSaveConfig()
	})

	th.SetThumbClickCallback(func() {
		st := v.model.SetCurrentStock(st.Symbol)
		v.SetChart(st)
		v.goSaveConfig()
	})
}

// Resize responds to window size changes by updating internal matrices.
func (v *View) Resize(newSize image.Point) {
	// Return if the window has not changed size.
	if v.winSize == newSize {
		return
	}

	gl.Viewport(0, 0, int32(newSize.X), int32(newSize.Y))

	v.winSize = newSize

	// Calculate the new ortho projection view matrix.
	fw, fh := float32(v.winSize.X), float32(v.winSize.Y)
	gfx.SetProjectionViewMatrix(math2.OrthoMatrix(fw, fh, fw /* use width as depth */))
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

func (v *View) ClearInputSymbol() {
	v.inputSymbol.Text = ""
}

func (v *View) SubmitSymbol() {
	st := v.model.SetCurrentStock(v.inputSymbol.Text)
	v.SetChart(st)
	v.GoRefreshStock(st)
	v.goSaveConfig()
	v.inputSymbol.Text = ""
}

// HandleCursorPos is a callback registered with GLFW to track cursor movement.
func (v *View) HandleCursorPos(x, y float64) {
	// Flip Y-axis since the OpenGL coordinate system makes lower left the origin.
	v.mousePos = image.Pt(int(x), v.winSize.Y-int(y))
}

// HandleMouseButton is a callback registered with GLFW to track mouse clicks.
func (v *View) HandleMouseButton(button glfw.MouseButton, action glfw.Action) {
	if button != glfw.MouseButtonLeft {
		glog.Infof("handleMouseButton: ignoring mouse button(%v) and action(%v)", button, action)
		return // Only interested in left clicks right now.
	}
	v.mouseLeftButtonClicked = action == glfw.Release
}

func (v *View) GoRefreshStock(st *ModelStock) {
	go func() {
		end := time2.Midnight(time.Now().In(time2.NewYorkLoc))
		start := end.Add(-6 * 30 * 24 * time.Hour)
		hist, err := stock.GetTradingHistory(&stock.GetTradingHistoryRequest{
			Symbol:    st.Symbol,
			StartDate: start,
			EndDate:   end,
		})
		if err != nil {
			glog.Warningf("goRefreshStock: failed to get trading history for %s: %v", st.Symbol, err)
			return
		}

		v.pendingStockUpdates <- viewStockUpdate{
			stock:          st,
			tradingHistory: hist,
		}
	}()
}

func (v *View) goSaveConfig() {
	// Make the config on the main thread to save the exact config at the time.
	cfg := &Config{}
	if st := v.model.CurrentStock; st != nil {
		cfg.CurrentStock = ConfigStock{st.Symbol}
	}
	for _, st := range v.model.SavedStocks {
		cfg.Stocks = append(cfg.Stocks, ConfigStock{st.Symbol})
	}

	// Handle saving to disk in a separate go routine.
	go func() {
		if err := SaveConfig(cfg); err != nil {
			glog.Warningf("goSaveConfig: failed to save config: %v", err)
		}
	}()
}
