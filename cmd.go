package broccli

import (
	"fmt"
	"log"
	"os"
	"path"
	"reflect"
	"sort"
	"text/tabwriter"
)

// Cmd represent a command which has a name (used in args when calling app), description, a handler that is called.
// Such command can have flags and arguments.  In addition to that, required environment variables can be set.
type Cmd struct {
	name      string
	desc      string
	flags     map[string]*param
	args      map[string]*param
	argsOrder []string
	argsIdx   int
	envVars   map[string]*param
	handler   func(*CLI) int
	options   cmdOptions
}

// AddFlag adds a flag to a command and returns a pointer to Param instance.
// Method requires name (eg. 'data' for '--data', alias (eg. 'd' for '-d'), placeholder for the value displayed on the
// 'help' screen, description, type of the value and additional validation that is set up with bit flags, eg. IsRequired
// or AllowMultipleValues.  If no additional flags are required, 0 should be used.
func (c *Cmd) AddFlag(n string, a string, hv string, d string, t int64, f int64, opts ...paramOption) {
	if c.flags == nil {
		c.flags = map[string]*param{}
	}
	c.flags[n] = &param{
		name:      n,
		alias:     a,
		desc:      d,
		helpValue: hv,
		valueType: t,
		flags:     f,
		options:   paramOptions{},
	}
	for _, o := range opts {
		o(&(c.flags[n].options))
	}
}

// AddArg adds an argument to a command and returns a pointer to Param instance.  It is the same as adding flag except
// it does not have an alias.
func (c *Cmd) AddArg(n string, hv string, d string, t int64, f int64, opts ...paramOption) {
	if c.argsIdx > 9 {
		log.Fatal("Only 10 arguments are allowed")
	}
	if c.args == nil {
		c.args = map[string]*param{}
	}
	c.args[n] = &param{
		name:      n,
		desc:      d,
		helpValue: hv,
		valueType: t,
		flags:     f,
		options:   paramOptions{},
	}
	if c.argsOrder == nil {
		c.argsOrder = make([]string, 10)
	}
	c.argsOrder[c.argsIdx] = n
	c.argsIdx++
	for _, o := range opts {
		o(&(c.args[n].options))
	}
}

// AddEnvVar adds a required environment variable to a command and returns a pointer to Param.  It's arguments are very
// similar to ones in previous AddArg and AddFlag methods.
func (c *Cmd) AddEnvVar(n string, d string, t int64, f int64, opts ...paramOption) {
	if c.envVars == nil {
		c.envVars = map[string]*param{}
	}
	c.envVars[n] = &param{
		name:      n,
		desc:      d,
		valueType: t,
		flags:     f,
		options:   paramOptions{},
	}
}

func (c *Cmd) sortedArgs() []string {
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

func (c *Cmd) sortedFlags() []string {
	fs := reflect.ValueOf(c.flags).MapKeys()
	sfs := make([]string, len(fs))
	for i, f := range fs {
		sfs[i] = f.String()
	}
	sort.Strings(sfs)
	return sfs
}

func (c *Cmd) sortedEnvVars() []string {
	evs := reflect.ValueOf(c.envVars).MapKeys()
	sevs := make([]string, len(evs))
	for i, ev := range evs {
		sevs[i] = ev.String()
	}
	sort.Strings(sevs)
	return sevs
}

// PrintHelp prints command usage information to stdout file.
func (c *Cmd) printHelp() {
	fmt.Fprintf(os.Stdout, fmt.Sprintf("\nUsage:  %s %s [FLAGS]%s\n\n", path.Base(os.Args[0]), c.name,
		c.argsHelpLine()))
	fmt.Fprintf(os.Stdout, fmt.Sprintf("%s\n", c.desc))

	if len(c.envVars) > 0 {
		fmt.Fprintf(os.Stdout, "\nRequired environment variables:\n")
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 8, 8, 0, '\t', 0)
		for _, n := range c.sortedEnvVars() {
			fmt.Fprintf(w, "%s\t%s\n", n, c.envVars[n].desc)
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

func (c *Cmd) argsHelpLine() string {
	sr := ""
	so := ""
	if c.argsIdx > 0 {
		for i := 0; i < c.argsIdx; i++ {
			n := c.argsOrder[i]
			arg := c.args[n]
			if arg.flags&IsRequired > 0 {
				sr += " " + arg.helpValue
			} else {
				so += " [" + arg.helpValue + "]"
			}
		}
	}
	return sr + so
}
