package main

import (
	"os"

	"github.com/mivinci/cli"
)

var root = cli.Command{
	Name:  "hello",
	Usage: "A helloworld command line app.",
	Args:  []string{"name"},
	Run: func(ctx *cli.Context) error {
		println("Hello,", ctx.Args()[0])
		return nil
	},
}

func main() {
	root.Exec(os.Args)
}
