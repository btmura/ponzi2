package config

import (
	"io/ioutil"
	"os"
	"os/user"
	"path"

	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
)

// Generate config.pb.go. Follow setup instructions @ github.com/golang/protobuf.
//go:generate protoc -I=data --go_out=. config.proto

// LoadConfig loads the user's config from disk.
func LoadConfig() (*Config, error) {
	cfgPath, err := userConfigPath()
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

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := proto.UnmarshalText(string(bytes), cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// SaveConfig saves the user's config to disk.
func SaveConfig(cfg *Config) error {
	cfgPath, err := userConfigPath()
	if err != nil {
		return err
	}

	glog.Infof("SaveConfig: saving config to %s", cfgPath)

	file, err := os.OpenFile(cfgPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0660)
	if err != nil {
		return err
	}
	defer file.Close()

	m := &proto.TextMarshaler{Compact: false}
	if err := m.Marshal(file, cfg); err != nil {
		return err
	}
	return nil
}

func userConfigPath() (string, error) {
	dirPath, err := userConfigDir()
	if err != nil {
		return "", err
	}
	return path.Join(dirPath, "config.txt"), nil
}

func userConfigDir() (string, error) {
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
