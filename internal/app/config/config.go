// Package config provides functions for loading and saving the user's configuration.
package config

import (
	"encoding/gob"
	"os"
	"os/user"
	"path"
	"path/filepath"

	"github.com/btmura/ponzi2/internal/app/model"
	"github.com/btmura/ponzi2/internal/app/view/chart"
	"github.com/btmura/ponzi2/internal/logger"
)

// Config configures the app.
type Config struct {
	CurrentStock *Stock
	Stocks       []*Stock
	Settings     Settings
}

// Stock identifies a single stock by symbol.
type Stock struct {
	Symbol string
}

// Settings has the user's settings.
type Settings struct {
	ChartSettings ChartSettings
}

// ChartSettings has the user's chart settings.
type ChartSettings struct {
	PriceStyle chart.PriceStyle
	Interval   model.Interval
}

// Load loads the user's config from disk.
func Load() (*Config, error) {
	cfgPath, err := userConfigPath()
	if err != nil {
		return nil, err
	}

	logger.Infof("loading from %s", cfgPath)

	file, err := os.Open(cfgPath)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Errorf("closing in load failed: %v", err)
		}
	}()

	cfg := &Config{}
	dec := gob.NewDecoder(file)
	if err := dec.Decode(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Save saves the user's config to disk.
func Save(cfg *Config) error {
	cfgPath, err := userConfigPath()
	if err != nil {
		return err
	}

	logger.Infof("saving to %s", cfgPath)

	file, err := os.OpenFile(cfgPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0660)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Errorf("closing in save failed: %v", err)
		}
	}()

	return gob.NewEncoder(file).Encode(cfg)
}

func userConfigPath() (string, error) {
	dirPath, err := userConfigDir()
	if err != nil {
		return "", err
	}
	return path.Join(dirPath, "config.gob"), nil
}

func userConfigDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	p := filepath.Join(u.HomeDir, ".config", "ponzi")
	if err := os.MkdirAll(p, 0755); err != nil {
		return "", err
	}
	return p, nil
}
