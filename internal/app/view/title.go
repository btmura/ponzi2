package view

import (
	"fmt"

	"github.com/go-gl/glfw/v3.2/glfw"

	"gitlab.com/btmura/ponzi2/internal/app/model"
)

// Title renders the text in the title bar.
type Title struct {
	text string
}

func NewTitle() *Title {
	return &Title{text: appName}
}

func (t *Title) SetData(st *model.Stock) {
	if st == nil {
		t.text = appName
		return
	}

	q := st.Quote
	if q == nil {
		t.text = appName
		return
	}

	t.text = fmt.Sprintf("%s - %s", st.Symbol, appName)
}

func (t *Title) Render(win *glfw.Window) {
	win.SetTitle(t.text)
}
