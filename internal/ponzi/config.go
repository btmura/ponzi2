package ponzi

import (
	"encoding/json"
	"io"
	"os"
	"os/user"
	"path"
	"sync"

	"github.com/golang/glog"
)

// Config has the user's saved stocks.
type Config struct {
	// CurrentStock is the stock the user is viewing.
	CurrentStock ConfigStock

	// Stocks are the config's stocks. Capitalized for JSON decoding.
	Stocks []ConfigStock
}

// ConfigStock represents a single user's stock.
type ConfigStock struct {
	// Symbol is the stock's symbol. Capitalized for JSON decoding.
	Symbol string
}

// configMutex prevents config file reads and writes from conflicting.
var configMutex sync.RWMutex

// LoadConfig loads the user's config from disk.
func LoadConfig() (*Config, error) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	cfgPath, err := getUserConfigPath()
	if err != nil {
		return nil, err
	}

	glog.Infof("LoadConfig: loading config from %s", cfgPath)

	file, err := os.Open(cfgPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	defer file.Close()

	if os.IsNotExist(err) {
		return &Config{}, nil
	}

	cfg := &Config{}
	d := json.NewDecoder(file)
	if err := d.Decode(&cfg); err != nil && err != io.EOF {
		return nil, err
	}
	return cfg, nil
}

// SaveConfig saves the user's config to disk.
func SaveConfig(cfg *Config) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	cfgPath, err := getUserConfigPath()
	if err != nil {
		return err
	}

	glog.Infof("SaveConfig: saving config to %s", cfgPath)

	file, err := os.OpenFile(cfgPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0660)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(&cfg)
}

func getUserConfigPath() (string, error) {
	dirPath, err := getUserConfigDir()
	if err != nil {
		return "", err
	}
	return path.Join(dirPath, "config.json"), nil
}

func getUserConfigDir() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	p := path.Join(u.HomeDir, ".config", "ponzi")
	if err := os.MkdirAll(p, 0755); err != nil {
		return "", err
	}
	return p, nil
}
