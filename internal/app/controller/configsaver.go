package controller

import (
	"github.com/golang/glog"

	"github.com/btmura/ponzi2/internal/app/config"
)

type configSaver struct {
	// queue is a queue of configs to save.
	queue chan *config.Config

	// done indicates saving is done and the program may quit.
	done chan bool

	// enabled enables saving configs when set to true.
	enabled bool
}

func newConfigSaver() *configSaver {
	return &configSaver{
		queue: make(chan *config.Config),
		done:  make(chan bool),
	}
}

func (c *configSaver) saveLoop() {
	for cfg := range c.queue {
		if err := config.Save(cfg); err != nil {
			glog.V(2).Infof("failed to save config: %v", err)
		}
	}
	c.done <- true
}

func (c *configSaver) start() {
	c.enabled = true
}

func (c *configSaver) save(cfg *config.Config) {
	if !c.enabled {
		glog.V(2).Infof("ignoring save request, saving disabled")
		return
	}

	// Queue the config for saving.
	go func() {
		c.queue <- cfg
	}()
}

func (c *configSaver) stop() {
	c.enabled = false
	close(c.queue)
	<-c.done
}
