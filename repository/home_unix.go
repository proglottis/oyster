// +build darwin linux

package repository

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
