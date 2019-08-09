// Package app exports a Run function to start the app.
package app

import (
	"github.com/btmura/ponzi2/internal/app/controller"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// App runs a GUI.
type App struct {
	client *iex.Client
}

// New returns a new App.
func New(client *iex.Client) *App {
	return &App{client}
}

// Run runs the app. Should be called from main.
func (a *App) Run() error {
	return controller.New(a.client).RunLoop()
}
