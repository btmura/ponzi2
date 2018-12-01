package view

import (
	"fmt"

	"github.com/go-gl/glfw/v3.2/glfw"
	"gitlab.com/btmura/ponzi2/internal/app/model"
)

// viewTitle renders the text in the title bar.
type viewTitle struct {
	title string
}

func newViewTitle() *viewTitle {
	return &viewTitle{title: appName}
}

func (t *viewTitle) SetData(st *model.Stock) {
	if st == nil {
		t.title = appName
		return
	}

	q := st.Quote
	if q == nil {
		t.title = appName
		return
	}

	t.title = fmt.Sprintf("%s - %s", st.Symbol, appName)
}

func (t *viewTitle) Render(win *glfw.Window) {
	win.SetTitle(t.title)
}
