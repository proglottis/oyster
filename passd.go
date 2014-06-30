package main

import (
	"os"
	"github.com/codegangsta/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "passd"
	app.Usage = "Password daemon"
	app.Run(os.Args)
}
