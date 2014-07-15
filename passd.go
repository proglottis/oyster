package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"os/user"
	"path"
	"time"
)

import (
	"github.com/atotto/clipboard"
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
		repo.Walk(func(file string) {
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
			Name:  "get",
			Usage: "Print a password to console",
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
			BashComplete: func(c *cli.Context) {
				if len(c.Args()) > 0 {
					return
				}
				repo.Walk(func(file string) {
					fmt.Println(file)
				})
			},
		},
		{
			Name:  "copy",
			Usage: "Copy a password to the clipboard",
			Action: func(c *cli.Context) {
				fmt.Printf("Password: ")
				passphrase := gopass.GetPasswd()
				password, err := repo.GetLine(c.Args().First(), passphrase)
				if err != nil {
					panic(err)
				}
				signals := make(chan os.Signal, 1)
				signal.Notify(signals, os.Interrupt, os.Kill)
				original, _ := clipboard.ReadAll()
				clipboard.WriteAll(password)
				select {
				case <-signals:
				case <-time.After(45 * time.Second):
				}
				current, _ := clipboard.ReadAll()
				if current == password {
					clipboard.WriteAll(original)
				}
			},
			BashComplete: func(c *cli.Context) {
				if len(c.Args()) > 0 {
					return
				}
				repo.Walk(func(file string) {
					fmt.Println(file)
				})
			},
		},
		{
			Name:  "put",
			Usage: "Store a password",
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
