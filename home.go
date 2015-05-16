package oyster

import (
	"os"
	"path"
)

func Home() string {
	home := os.Getenv("OYSTERHOME")
	if home == "" {
		home = path.Join(configDir(), hiddenPrefix+"oyster")
	}
	return home
}

func GpgHome() string {
	home := os.Getenv("GNUPGHOME")
	if home == "" {
		home = path.Join(configDir(), hiddenPrefix+"gnupg")
	}
	return home
}
