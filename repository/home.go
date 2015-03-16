package repository

import (
	"os"
	"os/user"
	"path"
)

func Home() string {
	home := os.Getenv("OYSTERHOME")
	if home == "" {
		user, err := user.Current()
		if err != nil {
			panic(err)
		}
		home = path.Join(user.HomeDir, ".oyster")
	}
	return home
}

func GpgHome() string {
	home := os.Getenv("GNUPGHOME")
	if home == "" {
		user, err := user.Current()
		if err != nil {
			panic(err)
		}
		home = path.Join(user.HomeDir, ".gnupg")
	}
	return home
}
