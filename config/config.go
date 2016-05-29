package config

import (
	"os"
	"path"

	"github.com/robfig/config"
)

type Config struct {
	ini *config.Config
}

func New() *Config {
	return &Config{ini: config.NewDefault()}
}

func Read() (*Config, error) {
	ini, err := config.ReadDefault(path.Join(configDir(), hiddenPrefix+"oysterconfig"))
	if err != nil {
		if os.IsNotExist(err) {
			return New(), nil
		}
		return nil, err
	}
	return &Config{ini: ini}, nil
}

func (c *Config) Home() string {
	var err error
	val := os.Getenv("OYSTERHOME")
	if val != "" {
		return val
	}
	val, err = c.ini.String("", "home")
	if err == nil {
		return val
	}
	return path.Join(configDir(), hiddenPrefix+"oyster")
}

func (c *Config) GpgHome() string {
	var err error
	val := os.Getenv("GNUPGHOME")
	if val != "" {
		return val
	}
	val, err = c.ini.String("", "gpgHome")
	if err == nil {
		return val
	}
	return path.Join(configDir(), hiddenPrefix+"gnupg")
}
