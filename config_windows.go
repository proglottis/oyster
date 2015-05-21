package oyster

import (
	"os"
)

const (
	hiddenPrefix = ""
)

func configDir() string {
	return os.Getenv("APPDATA")
}
