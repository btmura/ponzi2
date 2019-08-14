// Package app exports a Run function to start the app.
package app

import (
	"context"

	"github.com/btmura/ponzi2/internal/app/controller"
	"github.com/btmura/ponzi2/internal/errors"
	"github.com/btmura/ponzi2/internal/stock/iex"
)

// App runs a GUI.
type App struct {
	client iexClientInterface
	token  string
}

// iexClientInterface is implemented by clients in the iex package to get stock data.
type iexClientInterface interface {
	GetQuotes(ctx context.Context, req *iex.GetQuotesRequest) ([]*iex.Quote, error)
	GetCharts(ctx context.Context, req *iex.GetChartsRequest) ([]*iex.Chart, error)
}

// New returns a new App.
func New(client iexClientInterface, token string) *App {
	return &App{client, token}
}

// Run runs the app. Should be called from main.
func (a *App) Run() error {
	if a.client == nil {
		return errors.Errorf("nil client")
	}

	if a.token == "" {
		return errors.Errorf("missing token")
	}

	return controller.New(a.client, a.token).RunLoop()
}
