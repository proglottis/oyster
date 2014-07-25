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
	"code.google.com/p/go.crypto/ssh/terminal"
	"github.com/atotto/clipboard"
	"github.com/codegangsta/cli"
)

func repositoryHome() string {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	return path.Join(user.HomeDir, ".passd")
}

func copyThenClear(text string, d time.Duration) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	defer signal.Stop(signals)
	original, err := clipboard.ReadAll()
	if err != nil {
		return err
	}
	err = clipboard.WriteAll(text)
	if err != nil {
		return err
	}
	select {
	case <-signals:
	case <-time.After(d):
	}
	current, _ := clipboard.ReadAll()
	if current == text {
		return clipboard.WriteAll(original)
	}
	return nil
}

func bashCompleteKeys(repo Repository) func(*cli.Context) {
	return func(c *cli.Context) {
		if len(c.Args()) > 0 {
			return
		}
		repo.Walk(func(file string) {
			fmt.Println(file)
		})
	}
}

type password struct {
	Password []byte
	Err      error
}

func interruptibleCopy(dst io.Writer, src io.Reader) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
	defer signal.Stop(signals)
	done := make(chan struct{})
	go func() {
		io.Copy(dst, src)
		close(done)
	}()
	select {
	case <-signals:
	case <-done:
	}
}

func getPassword() ([]byte, error) {
	signals := make(chan os.Signal, 1)
	passwords := make(chan password)
	signal.Notify(signals, os.Interrupt, os.Kill)
	defer signal.Stop(signals)
	state, err := terminal.GetState(0)
	if err != nil {
		return nil, err
	}
	defer terminal.Restore(0, state)
	go func() {
		fmt.Printf("Password: ")
		defer fmt.Printf("\n")
		p, err := terminal.ReadPassword(0)
		passwords <- password{
			Password: p,
			Err:      err,
		}
		close(passwords)
	}()
	select {
	case <-signals:
		return nil, fmt.Errorf("Password entry cancelled")
	case password := <-passwords:
		return password.Password, password.Err
	}
	panic("unreachable")
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
				args := c.Args()
				if !args.Present() {
					fmt.Println("Must provide at least one GPG ID")
					return
				}
				if err := repo.Init(args); err != nil {
					fmt.Println(err)
				}
			},
		},
		{
			Name:  "get",
			Usage: "Print a password to console",
			Action: func(c *cli.Context) {
				passphrase, err := getPassword()
				if err != nil {
					panic(err)
				}
				plaintext, err := repo.Get(c.Args().First(), passphrase)
				if err != nil {
					panic(err)
				}
				defer plaintext.Close()
				io.Copy(os.Stdout, plaintext)
			},
			BashComplete: bashCompleteKeys(repo),
		},
		{
			Name:  "copy",
			Usage: "Copy a password to the clipboard",
			Action: func(c *cli.Context) {
				passphrase, err := getPassword()
				if err != nil {
					panic(err)
				}
				password, err := repo.GetLine(c.Args().First(), passphrase)
				if err != nil {
					panic(err)
				}
				err = copyThenClear(password, 45*time.Second)
				if err != nil {
					panic(err)
				}
			},
			BashComplete: bashCompleteKeys(repo),
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
				if terminal.IsTerminal(0) {
					fmt.Println("Enter your password...")
				}
				interruptibleCopy(plaintext, os.Stdin)
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
			BashComplete: bashCompleteKeys(repo),
		},
		{
			Name:  "server",
			Usage: "Start password daemon",
			Action: func(c *cli.Context) {
				RunServer(repo)
			},
		},
	}
	app.Run(os.Args)
}
