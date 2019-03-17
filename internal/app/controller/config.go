package controller

import (
	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/app/config"
	"github.com/btmura/ponzi2/internal/app/model"
)

type configController struct {
	// enableSavingConfigs enables saving config changes.
	enableSavingConfigs bool

	// pendingConfigSaves is a channel with configs to save.
	pendingConfigSaves chan *config.Config

	// doneSavingConfigs indicates saving is done and the program may quit.
	doneSavingConfigs chan bool
}

func newConfigController() *configController {
	return &configController{
		pendingConfigSaves: make(chan *config.Config),
		doneSavingConfigs:  make(chan bool),
	}
}

func (c *configController) saveLoop() {
	for cfg := range c.pendingConfigSaves {
		if err := config.Save(cfg); err != nil {
			glog.V(2).Infof("failed to save config: %v", err)
		}
	}
	c.doneSavingConfigs <- true
}

func (c *configController) start() {
	c.enableSavingConfigs = true
}

func (c *configController) save(model *model.Model) {
	if !c.enableSavingConfigs {
		glog.V(2).Infof("ignoring save request, saving disabled")
		return
	}

	// Make the config on the main thread to save the exact config at the time.
	cfg := &config.Config{}
	if s := model.CurrentSymbol(); s != "" {
		cfg.CurrentStock = &config.Stock{Symbol: s}
	}
	for _, s := range model.SidebarSymbols() {
		cfg.Stocks = append(cfg.Stocks, &config.Stock{Symbol: s})
	}

	// Queue the config for saving.
	go func() {
		c.pendingConfigSaves <- cfg
	}()
}

func (c *configController) stop() {
	c.enableSavingConfigs = false
	close(c.pendingConfigSaves)
	<-c.doneSavingConfigs
}
