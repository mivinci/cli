package cli

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

var ErrHelp = errors.New("help requested")

type Context struct {
	out  io.Writer
	cmd  *Command
	args []string // non-flag arguments
	lset map[string]*Flag
	sset map[byte]*Flag
}

func (c *Context) String(key string) string {
	flag, ok := c.lset[key]
	if !ok {
		return ""
	}
	return flag.Value.String()
}

func (c *Context) Int(key string) int {
	flag, ok := c.lset[key]
	if !ok {
		return 0
	}
	iv, ok := flag.Value.(*intValue)
	if !ok {
		return 0
	}
	return int(*iv)
}

func (c *Context) Bool(key string) bool {
	flag, ok := c.lset[key]
	if !ok {
		return false
	}
	bv, ok := flag.Value.(*boolValue)
	if !ok {
		return false
	}
	return bool(*bv)
}

func (c *Context) IP(key string) net.IP {
	flag, ok := c.lset[key]
	if !ok {
		return nil
	}
	ip, ok := flag.Value.(*ipValue)
	if !ok {
		return nil
	}
	return net.IP(*ip)
}

func (c *Context) Command() *Command {
	return c.cmd
}

func (c *Context) add(flag *Flag) {
	value := flag.Value
	if value != nil && value.Type() == "bool" {
		flag.optional = true
	}

	name := flag.Name
	if _, ok := c.lset[name]; ok {
		panic(fmt.Sprintf("flag %s redefined", name))
	}
	c.lset[name] = flag

	short := flag.Short
	if short == 0 {
		return
	}
	if used, ok := c.sset[short]; ok {
		panic(fmt.Sprintf("short flag %c is used by flag %s", short, used.Name))
	}
	c.sset[short] = flag
}

func (c *Context) Get(name string) (*Flag, bool) {
	flag, ok := c.lset[name]
	return flag, ok
}

func (c *Context) Args() []string {
	return c.args
}

func (c *Context) Foreach(fn func(*Flag)) {
	for _, flag := range c.lset {
		fn(flag)
	}
}

func (c *Context) ForeachPersist(fn func(*Flag)) {
	for _, flag := range c.lset {
		if flag.Persist {
			fn(flag)
		}
	}
}

func (c *Context) Parse(args []string) error {
	if err := c.parse(args); err != nil {
		if c.cmd.Options.enabled(ExitOnError) {
			fmt.Fprint(c.out, err)
			os.Exit(2)
		}
		return err
	}
	return nil
}

func (c *Context) parse(args []string) (err error) {
	c.args = make([]string, 0, len(args))

	for len(args) > 0 {
		arg := args[0]
		args = args[1:]
		if len(arg) == 0 || len(arg) == 1 || arg[0] != '-' {
			c.args = append(c.args, arg)
			continue
		}
		if arg[1] == '-' {
			if len(arg) == 2 { // `--` terminates the flags
				c.args = append(c.args, args...)
				break
			}
			args, err = c.parseLongArgs(arg, args)
		} else {
			args, err = c.parseShortArgs(arg, args)
		}
		if err != nil {
			return
		}
	}
	return
}

func (c *Context) parseLongArgs(arg string, args []string) ([]string, error) {
	var key, value string
	name := arg[2:]
	if len(name) == 0 || name[0] == '-' || name[0] == '=' {
		return nil, fmt.Errorf("unexpected flag %s", arg)
	}

	key = name
	i := strings.IndexByte(key, '=')
	if i > -1 {
		key = name[:i]
	}

	flag, ok := c.lset[key]
	if !ok {
		switch {
		case key == "help":
			c.cmd.Help()
			return args, ErrHelp
		case c.cmd.Options.enabled(UndefinedFlags):
			return stripUndefinedFlagValue(args), nil
		default:
			return nil, fmt.Errorf("undefined flag: --%s", key)
		}
	}
	if key == "help" {
		c.cmd.Help()
		return args, ErrHelp
	}

	switch {
	case i > -1:
		// --key=value
		value = name[i:]
	case flag.optional:
		// --key
		value = "true"
	case len(args) > 0:
		// --key value
		value = args[0]
		args = args[1:]
	default:
		return nil, fmt.Errorf("flag %s needs an argument", key)
	}

	return args, flag.Value.Set(value)
}

func (c *Context) parseShortArgs(arg string, args []string) ([]string, error) {
	var value string
	keys := arg[1:]
	for len(keys) > 0 {
		ch := keys[0]
		keys = keys[1:]
		flag, ok := c.sset[ch]
		if !ok {
			switch {
			case ch == 'h':
				c.cmd.Help()
				return nil, ErrHelp
			case c.cmd.Options.enabled(UndefinedFlags):
				return stripUndefinedFlagValue(args), nil
			default:
				return nil, fmt.Errorf("undefined shorthand flag: -%c", ch)
			}
		}
		if ch == 'h' {
			c.cmd.Help()
			return nil, ErrHelp
		}

		switch {
		case len(keys) > 2 && keys[1] == '=':
			// -k=value
			value = keys[2:]
		case flag.optional:
			// -k
			value = "true"
		case len(keys) > 1:
			// -kvalue
			value = keys[1:]
		case len(args) > 0:
			// -k value
			value = args[0]
			args = args[1:]
		default:
			return nil, fmt.Errorf("flag %c needs an argument", ch)
		}

		if err := flag.Value.Set(value); err != nil {
			return nil, err
		}
	}
	return args, nil
}

func stripUndefinedFlagValue(args []string) []string {
	// --undefined
	if len(args) == 0 {
		return args
	}
	// --undefined --next-key
	next := args[0]
	if len(next) > 0 && next[0] == '-' {
		return args
	}
	// --undefined arg ...
	if len(args) > 1 {
		return args[1:]
	}
	return nil
}
