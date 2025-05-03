package broccli

import (
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"reflect"
	"sort"
	"text/tabwriter"
)

// Command represent a command which has a name (used in args when calling app), usage, a handler that is called.
// Such command can have flags and arguments.  In addition to that, required environment variables can be set.
type Command struct {
	name      string
	usage      string
	flags     map[string]*param
	args      map[string]*param
	argsOrder []string
	argsIdx   int
	env   map[string]*param
	handler   func(context.Context, *Broccli) int
	options   commandOptions
}

// Flag adds a flag to a command and returns a pointer to Param instance.
// Method requires name (eg. 'data' for '--data', alias (eg. 'd' for '-d'), placeholder for the value displayed on the
// 'help' screen, usage, type of the value and additional validation that is set up with bit flags, eg. IsRequired
// or AllowMultipleValues.  If no additional flags are required, 0 should be used.
func (c *Command) Flag(name, alias, valuePlaceholder, usage string, types, flags int64, opts ...paramOption) {
	if c.flags == nil {
		c.flags = map[string]*param{}
	}
	c.flags[name] = &param{
		name:             name,
		alias:            alias,
		usage:             usage,
		valuePlaceholder: valuePlaceholder,
		valueType:        types,
		flags:            flags,
		options:          paramOptions{},
	}
	for _, o := range opts {
		o(&(c.flags[name].options))
	}
}

// Arg adds an argument to a command and returns a pointer to Param instance.  It is the same as adding flag except
// it does not have an alias.
func (c *Command) Arg(name, valuePlaceholder, usage string, types, flags int64, opts ...paramOption) {
	if c.argsIdx > 9 {
		log.Fatal("Only 10 arguments are allowed")
	}
	if c.args == nil {
		c.args = map[string]*param{}
	}
	c.args[name] = &param{
		name:             name,
		usage:             usage,
		valuePlaceholder: valuePlaceholder,
		valueType:        types,
		flags:            flags,
		options:          paramOptions{},
	}
	if c.argsOrder == nil {
		c.argsOrder = make([]string, 10)
	}
	c.argsOrder[c.argsIdx] = name
	c.argsIdx++
	for _, o := range opts {
		o(&(c.args[name].options))
	}
}

// Env adds a required environment variable to a command and returns a pointer to Param.  It's arguments are very
// similar to ones in previous AddArg and AddFlag methods.
func (c *Command) Env(name, usage string, types, flags int64, opts ...paramOption) {
	if c.env == nil {
		c.env = map[string]*param{}
	}
	c.env[name] = &param{
		name:      name,
		usage:      usage,
		valueType: types,
		flags:     flags,
		options:   paramOptions{},
	}
}

func (c *Command) sortedArgs() []string {
	args := make([]string, c.argsIdx)
	idx := 0
	for i := 0; i < c.argsIdx; i++ {
		n := c.argsOrder[i]
		arg := c.args[n]
		if arg.flags&IsRequired > 0 {
			args[idx] = n
			idx++
		}
	}
	for i := 0; i < c.argsIdx; i++ {
		n := c.argsOrder[i]
		arg := c.args[n]
		if arg.flags&IsRequired == 0 {
			args[idx] = n
			idx++
		}
	}
	return args
}

func (c *Command) sortedFlags() []string {
	fs := reflect.ValueOf(c.flags).MapKeys()
	sfs := make([]string, len(fs))
	for i, f := range fs {
		sfs[i] = f.String()
	}
	sort.Strings(sfs)
	return sfs
}

func (c *Command) sortedEnv() []string {
	evs := reflect.ValueOf(c.env).MapKeys()
	sevs := make([]string, len(evs))
	for i, ev := range evs {
		sevs[i] = ev.String()
	}
	sort.Strings(sevs)
	return sevs
}

// PrintHelp prints command usage information to stdout file.
func (c *Command) printHelp() {
	fmt.Fprintf(os.Stdout, "\nUsage:  %s %s [FLAGS]%s\n\n", path.Base(os.Args[0]), c.name,
		c.argsHelpLine())
	fmt.Fprintf(os.Stdout, "%s\n", c.usage)

	if len(c.env) > 0 {
		fmt.Fprintf(os.Stdout, "\nRequired environment variables:\n")
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 8, 8, 0, '\t', 0)
		for _, n := range c.sortedEnv() {
			fmt.Fprintf(w, "%s\t%s\n", n, c.env[n].usage)
		}
		w.Flush()
	}

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 8, 8, 0, '\t', 0)

	var s [2]string
	i := 1
	for _, n := range c.sortedFlags() {
		flag := c.flags[n]
		if flag.flags&IsRequired > 0 {
			i = 0
		} else {
			i = 1
		}
		s[i] += flag.helpLine()
	}

	if s[0] != "" {
		fmt.Fprintf(w, "\nRequired flags: \n")
		fmt.Fprintf(w, s[0])
		w.Flush()
	}
	if s[1] != "" {
		fmt.Fprintf(w, "\nOptional flags: \n")
		fmt.Fprintf(w, s[1])
		w.Flush()
	}

}

func (c *Command) argsHelpLine() string {
	sr := ""
	so := ""
	if c.argsIdx > 0 {
		for i := 0; i < c.argsIdx; i++ {
			n := c.argsOrder[i]
			arg := c.args[n]
			if arg.flags&IsRequired > 0 {
				sr += " " + arg.valuePlaceholder
			} else {
				so += " [" + arg.valuePlaceholder + "]"
			}
		}
	}
	return sr + so
}
