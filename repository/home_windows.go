package repository

import (
	"os"
)

const (
	hiddenPrefix = ""
)

func configDir() string {
	return os.Getenv("APPDATA")
}
