// +build darwin linux

package oyster

import (
	"os/user"
)

const (
	hiddenPrefix = "."
)

func configDir() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	return user.HomeDir
}
