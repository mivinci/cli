package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mivinci/cli"
)

func main() {
	root := cli.Command{
		Name:  "example",
		Usage: `A command line app demonstrating subcommands.`,
		Flags: []*cli.Flag{
			{
				Name:  "address",
				Usage: "specify an IP address",
				Value: cli.IP("127.0.0.1"),
			},
			{
				Name:  "hash",
				Usage: "specify a hash number",
				Value: cli.Int(53748191),
			},
			{
				Name:  "verbose",
				Usage: "show tails when running",
				Value: cli.Bool(false),
			},
		},
		Run: func(ctx *cli.Context) error {
			fmt.Println(ctx.IP("address"))
			return nil
		},
	}
	upload := cli.Command{
		Name:  "upload",
		Usage: "upload a file",
		Run: func(ctx *cli.Context) error {
			fmt.Println("a1")
			fmt.Println(ctx.Args())
			fmt.Println(ctx.Bool("foo"))
			return nil
		},
	}
	download := cli.Command{
		Name:  "download",
		Usage: "download a file",
		Flags: []*cli.Flag{
			{
				Name:  "thread",
				Short: 't',
				Usage: "specify the number of threads",
				Value: cli.Int(2),
			},
		},
		Run: func(ctx *cli.Context) error {
			fmt.Println(ctx.Args())
			fmt.Println(ctx.Bool("foo"))
			return nil
		},
	}

	root.Add(&upload, &download)

	if err := root.Exec(os.Args); err != nil {
		log.Fatal(err)
	}
}
