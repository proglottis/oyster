package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"time"

	"github.com/atotto/clipboard"
	"github.com/codegangsta/cli"
	"github.com/proglottis/oyster"
	"github.com/sourcegraph/rwvfs"
	"golang.org/x/crypto/ssh/terminal"
)

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

func bashCompleteKeys(repo *oyster.FileRepo) func(*cli.Context) {
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
	state, err := terminal.GetState(2)
	if err != nil {
		return nil, err
	}
	defer terminal.Restore(2, state)
	go func() {
		fmt.Fprintf(os.Stderr, "Password: ")
		defer fmt.Fprintf(os.Stderr, "\n")
		p, err := terminal.ReadPassword(2)
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
}

func main() {
	gpg := oyster.NewGpgRepo(oyster.GpgHome())
	fs := oyster.NewCryptoFS(rwvfs.OSPerm(oyster.Home(), 0600, 0700), gpg)
	repo := oyster.NewFileRepo(fs)
	app := cli.NewApp()
	app.Name = "oyster"
	app.Usage = "PGP password storage"
	app.Version = "0.2.4"
	app.EnableBashCompletion = true
	app.Action = func(c *cli.Context) {
		repo.Walk(func(file string) {
			fmt.Println(file)
		})
	}

	cli.CommandHelpTemplate = `NAME:
  {{.Name}} - {{.Usage}}

USAGE:
  oyster {{.Name}}{{if .Flags}} [command options]{{end}} [arguments...]{{if .Description}}

DESCRIPTION:
  {{.Description}}{{end}}{{if .Flags}}

OPTIONS:
  {{range .Flags}}{{.}}
  {{end}}{{ end }}
`

	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Setup Oyster",
			Description: `Create Oyster home directory. If OYSTERHOME is set it will be used instead of "~/.oyster".

EXAMPLE:
   oyster init me@example.org
`,
			Action: func(c *cli.Context) {
				args := c.Args()
				if !args.Present() {
					fmt.Println("Must provide at least one GPG ID")
					return
				}
				if err := oyster.InitRepo(fs, args); err != nil {
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
				plaintext, err := repo.Open(c.Args().First(), passphrase)
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
				password, err := repo.Line(c.Args().First(), passphrase)
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
				plaintext, err := repo.Create(c.Args().First())
				if err != nil {
					panic(err)
				}
				defer plaintext.Close()
				if terminal.IsTerminal(0) {
					fmt.Println("Enter your password...")
				}
				interruptibleCopy(plaintext, os.Stdin)
			},
			BashComplete: bashCompleteKeys(repo),
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
	}
	app.Run(os.Args)
}
