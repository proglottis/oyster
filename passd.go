package main

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"path"
)

import (
	"github.com/codegangsta/cli"
	"github.com/howeyc/gopass"
)

func repositoryHome() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	return path.Join(user.HomeDir, ".passd")
}

func main() {
	repo := NewRepository(repositoryHome())
	app := cli.NewApp()
	app.Name = "passd"
	app.Usage = "Password daemon"
	app.Commands = []cli.Command{
		{
			Name:      "get",
			ShortName: "g",
			Usage:     "Print a password",
			Action: func(c *cli.Context) {
				fmt.Printf("Password: ")
				passphrase := gopass.GetPasswd()
				plaintext, err := repo.Get(c.Args().First(), passphrase)
				if err != nil {
					panic(err)
				}
				defer plaintext.Close()
				io.Copy(os.Stdout, plaintext)
			},
		},
		{
			Name:      "put",
			ShortName: "p",
			Usage:     "Store a password",
			Action: func(c *cli.Context) {
				plaintext, err := repo.Put(c.Args().First())
				if err != nil {
					panic(err)
				}
				defer plaintext.Close()
				io.Copy(plaintext, os.Stdin)
			},
		},
		{
			Name:      "remove",
			ShortName: "rm",
			Usage:     "Remove a password",
			Action: func(c *cli.Context) {
				if err := repo.Remove(c.Args().First()); err != nil {
					panic(err)
				}
			},
		},
	}
	app.Run(os.Args)
}
