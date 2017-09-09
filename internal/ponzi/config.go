package ponzi

package ponzi

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"os/user"
	"path"
	"sync"
)

// config has the user's saved stocks.
type config struct {
	// Stocks are the config's stocks. Capitalized for JSON decoding.
	Stocks []configStock
}

// configStock represents a single user's stock.
type configStock struct {
	// Symbol is the stock's symbol. Capitalized for JSON decoding.
	Symbol string
}

// configMutex prevents config file reads and writes from conflicting.
var configMutex sync.RWMutex

// loadConfig loads the user's config from disk.
func loadConfig() (config, error) {
	configMutex.RLock()
	defer configMutex.RUnlock()

	cfgPath, err := getUserConfigPath()
	if err != nil {
		return config{}, err
	}

	file, err := os.Open(cfgPath)
	if err != nil && !os.IsNotExist(err) {
		return config{}, err
	}
	defer file.Close()

	if os.IsNotExist(err) {
		return config{}, nil
	}

	cfg := config{}
	d := json.NewDecoder(file)
	if err := d.Decode(&cfg); err != nil && err != io.EOF {
		return config{}, err
	}
	return cfg, nil
}

// saveConfig saves the user's config to disk.
func saveConfig(cfg config) error {
	configMutex.Lock()
	defer configMutex.Unlock()

	cfgPath, err := getUserConfigPath()
	if err != nil {
		return err
	}

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
