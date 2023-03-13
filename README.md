# cli

A package for building command line apps. It looks like [spf13/cobra](https://github.com/spf13/cobra) but has [urfave/cli](https://github.com/urfave/cli)-like APIs and a builtin argument parser also.


## Features

- POSIX-like syntax
- Nested commands
- Positional arguments
- Type-safe flag values
- Minimal (no duplicated flags defined with std/flag)
- Lightweight (no other packages are needed)
- Pluggable config file watcher (TODO)

## Supported flag types

- int, uint
- bool
- string
- time.Duration
- net.IP

## Usage

### Installation

```bash
go get -u github.com/mivinci/cli
```

### Quickstart

Here's a simple helloworld example.

```go
import (
    "os"

    "github.com/mivinci/cli"
)

var cmd = cli.Command{
    Name:  "hello",
    Usage: "A helloworld command line app.",
    Args:  []string{"name"},
    Run: func(ctx *cli.Context) error {
        println("Hello,", ctx.Args()[0])
        return nil
    },
}

func main() {
    cmd.Exec(os.Args)
}
```

Compile and run with `-h`, the output will be as follows.

```
$ go run hello.go -h
A helloworld command line app.

Usage:
  hello <name> [option]*

Available options:
  -h, --help    show help message

Use "hello [command] --help" for more infomation about a command.
```

Checkout [example](./example) for details.

## Feedbacks

- [Issues](https://github.com/mivinci/cli/issues)

## License

mivinci/cli is Apache 2.0 licensed.
