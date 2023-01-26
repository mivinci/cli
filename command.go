package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	UndefinedFlags Options = 1 << iota
	ExitOnError
)

type Options uint8

func (o Options) enabled(opts Options) bool {
	return o&opts != 0
}

var (
	defaultFlagHelp = Flag{
		Name:  "help",
		Short: 'h',
		Usage: "show help message",
		Value: Bool(false),
	}
)

type Flag struct {
	Name    string
	Usage   string
	Short   byte
	Value   Value
	Persist bool

	optional bool
}

type Command struct {
	Name    string
	Version string
	Example string
	Usage   string
	Args    []string // positional arguments
	Flags   []*Flag  // flag arguments
	Options Options
	Run     func(*Context) error

	children []*Command
	parent   *Command
	setup    bool
}

func (c *Command) Root() *Command {
	if c.parent == nil {
		return c
	}
	return c.parent.Root()
}

func (c *Command) Path() string {
	var path string
	for p := c; p != nil; p = p.parent {
		path = p.Name + " " + path
	}
	return path[:len(path)-1]
}

func (c *Command) Add(cmds ...*Command) {
	for i, cmd := range cmds {
		if cmds[i] == c {
			panic("command cannot be a child of itself")
		}
		cmds[i].parent = c
		c.children = append(c.children, cmd)
	}
}

func (c *Command) Exec(args []string) error {
	return c.ExecContext(context.Background(), args)
}

func (c *Command) ExecContext(ctx context.Context, args []string) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if c.parent != nil {
		return c.Root().ExecContext(ctx, args)
	}

	if !c.setup {
		c.Flags = append(c.Flags, &defaultFlagHelp)
		c.setup = true
	}

	args[0] = c.Name // always make the first argument match the root command
	cmd, args, err := c.Find(args)
	if err != nil {
		return err
	}

	cctx := newContext(cmd)
	if err = cctx.Parse(args); err != nil {
		if errors.Is(err, ErrHelp) {
			return nil
		}
		return err
	}

	if cmd.Run != nil {
		return cmd.Run(cctx)
	}
	return nil
}

// Find finds the nearest possible command matching the given arguments.
func (c *Command) Find(args []string) (*Command, []string, error) {
	var dummy Command
	dummy.children = []*Command{c}
	return find(&dummy, args)
}

func find(root *Command, args []string) (*Command, []string, error) {
	if len(args) == 0 {
		return root, args, nil
	}

	flags := make([]string, 0, len(args))

LOOP:
	for len(args) > 0 {
		arg := args[0]
		args = args[1:]

		switch {
		case arg == "--": // `--` terminates the arguments
			break LOOP
		case arg[0] == '-': // we met a flag
			flags = append(flags, arg)
		default:
			for _, cmd := range root.children { // see if it matches a subcommand
				if arg == cmd.Name {
					root = cmd
					continue LOOP
				}
			}
			flags = append(flags, arg) // oh it's just a value for a flag
		}
	}

	return root, flags, nil
}

func newContext(cmd *Command) *Context {
	ctx := &Context{
		cmd:  cmd,
		lset: make(map[string]*Flag),
		sset: make(map[byte]*Flag),
		out:  os.Stdout,
	}

	for p := cmd; p != nil; p = p.parent {
		for _, flag := range p.Flags {
			// always add root flags
			if p.parent == nil || flag.Persist {
				ctx.add(flag)
			}
		}
	}
	return ctx
}

func (c *Command) Help() {
	buf := bytes.Buffer{}
	path := c.Path()

	fmt.Fprintf(&buf, "%s\n\n", c.Usage)
	fmt.Fprintf(&buf, "Usage:\n  %s ", path)
	for _, arg := range c.Args {
		fmt.Fprintf(&buf, "<%s> ", arg)
	}
	if len(c.children) > 0 {
		fmt.Fprint(&buf, "[command] ")
	}
	if len(c.Flags) > 0 {
		fmt.Fprintln(&buf, "[option]*")
	}

	if len(c.Example) > 0 {
		fmt.Fprintf(&buf, "\nExample:\n  %s\n", c.Example)
	}

	if len(c.children) > 0 {
		fmt.Fprintln(&buf, "\nAvailable commands:")
		helpSubCommands(&buf, c.children)
	}

	if len(c.Flags) > 0 {
		fmt.Fprintln(&buf, "\nAvailable options:")
		helpFlags(&buf, c.Flags)
	}

	if c.parent != nil {
		fmt.Fprintln(&buf, "\nGlobal options:")
	}
	for p := c.parent; p != nil; p = p.parent {
		helpFlags(&buf, p.Flags)
	}

	fmt.Fprintf(&buf, "\nUse \"%s [command] --help\" for more infomation about a command.", path)
	fmt.Fprintln(os.Stdout, buf.String())
}

func helpSubCommands(buf *bytes.Buffer, cmds []*Command) {
	lines := make([]string, 0, len(cmds))
	maxlen := 0

	for _, cmd := range cmds {
		line := fmt.Sprintf("  %s\x00 ", cmd.Name)
		if len(line) > maxlen {
			maxlen = len(line)
		}
		line += cmd.Usage
		lines = append(lines, line)
	}

	for _, line := range lines {
		sidx := strings.Index(line, "\x00")
		spacing := strings.Repeat(" ", maxlen-sidx)
		fmt.Fprintln(buf, line[:sidx], spacing, line[sidx+1:])
	}
}

func helpFlags(buf *bytes.Buffer, flags []*Flag) {
	lines := make([]string, 0, len(flags))
	maxlen := 0

	for _, flag := range flags {
		var line string
		if flag.Short != 0 {
			line = fmt.Sprintf("  -%c, --%s ", flag.Short, flag.Name)
		} else {
			line = fmt.Sprintf("      --%s ", flag.Name)
		}
		value := flag.Value
		if value != nil && value.Type() != "bool" {
			line += value.Type()
		}
		line += "\x00"
		if len(line) > maxlen {
			maxlen = len(line)
		}
		line += flag.Usage

		if value != nil {
			def := flag.Value.String()
			if !flag.optional && len(def) > 0 {
				line += fmt.Sprintf(" (default %s)", def)
			}
		}
		lines = append(lines, line)
	}

	for _, line := range lines {
		sidx := strings.Index(line, "\x00")
		spacing := strings.Repeat(" ", maxlen-sidx)
		fmt.Fprintln(buf, line[:sidx], spacing, line[sidx+1:])
	}
}
