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
	app.EnableBashCompletion = true
	app.Action = func(c *cli.Context) {
		if len(c.Args()) > 0 {
			fmt.Printf("Password: ")
			passphrase := gopass.GetPasswd()
			plaintext, err := repo.Get(c.Args().First(), passphrase)
			if err != nil {
				panic(err)
			}
			defer plaintext.Close()
			io.Copy(os.Stdout, plaintext)
		} else {
			repo.Walk(func(file string){
				fmt.Println(file)
			})
		}
	}
	app.BashComplete = func(c *cli.Context) {
		if len(c.Args()) > 0 {
			return
		}
		repo.Walk(func(file string){
			fmt.Println(file)
		})
	}
	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Setup passd",
			Action: func(c *cli.Context) {
				if err := repo.Init(c.Args()); err != nil {
					panic(err)
				}
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
