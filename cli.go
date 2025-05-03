package broccli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"text/tabwriter"
)

// Broccli is main CLI application definition.
// It has a name, description, author which are printed out to the screen in the usage syntax.
// Each CLI have commands (represented by Command).  Optionally, it is possible to require environment
// variables.
type Broccli struct {
	name        string
	usage string
	author      string
	commands        map[string]*Command
	env     map[string]*param
	parsedFlags map[string]string
	parsedArgs  map[string]string
}

// NewBroccli returns pointer to a new Broccli instance.  Name, usage and author are displayed on the syntax screen.
func NewBroccli(name, usage, author string) *Broccli {
	c := &Broccli{
		name:        name,
		usage: usage,
		author:      author,
		commands:        map[string]*Command{},
		env:     map[string]*param{},
		parsedFlags: map[string]string{},
		parsedArgs:  map[string]string{},
	}
	return c
}

// Command returns pointer to a new command with specified name, usage and handler.  Handler is a function that
// gets called when command is executed.
// Additionally, there is a set of options that can be passed as arguments.  Search for commandOption for more info.
func (c *Broccli) Command(name, usage string, handler func(ctx context.Context, cli *Broccli) int, opts ...commandOption) *Command {
	c.commands[name] = &Command{
		name:    name,
		usage:    usage,
		flags:   map[string]*param{},
		args:    map[string]*param{},
		env: map[string]*param{},
		handler: handler,
		options: commandOptions{},
	}
	for _, o := range opts {
		o(&(c.commands[name].options))
	}
	return c.commands[name]
}

// Env returns pointer to a new environment variable that is required to run every command.
// Method requires name, eg. MY_VAR, and usage.
func (c *Broccli) Env(name string, usage string) {
	c.env[name] = &param{
		name:    name,
		usage:    usage,
		flags:   IsRequired,
		options: paramOptions{},
	}
}

// Flag returns value of flag.
func (c *Broccli) Flag(name string) string {
	return c.parsedFlags[name]
}

// Arg returns value of arg.
func (c *Broccli) Arg(name string) string {
	return c.parsedArgs[name]
}

// Run parses the arguments, validates them and executes command handler.
// In case of invalid arguments, error is printed to stderr and 1 is returned.  Return value should be treated as exit
// code.
func (c *Broccli) Run(ctx context.Context) int {
	// display help, first arg is binary filename
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		c.printHelp()
		return 0
	}
	for _, n := range c.sortedCommands() {
		if n != os.Args[1] {
			continue
		}
		// display command help
		if len(os.Args) > 2 && (os.Args[2] == "-h" || os.Args[2] == "--help") {
			c.commands[n].printHelp()
			return 0
		}

		// check required environment variables
		if len(c.env) > 0 {
			for env, param := range c.env {
				v := os.Getenv(env)
				param.flags = param.flags | IsRequired
				err := param.validateValue(v)
				if err != nil {
					fmt.Fprintf(os.Stderr, "ERROR: %s %s: %s\n", c.getParamTypeName(ParamEnvVar), param.name, err.Error())
					c.printHelp()
					return 1
				}
			}
		}

		// parse and validate all the flags and args
		exitCode := c.parseFlags(c.commands[n])
		if exitCode > 0 {
			return exitCode
		}

		return c.commands[n].handler(ctx, c)
	}

	// command not found
	c.printInvalidCommand(os.Args[1])
	return 1
}

func (c *Broccli) sortedCommands() []string {
	cmds := reflect.ValueOf(c.commands).MapKeys()
	scmds := make([]string, len(cmds))
	for i, cmd := range cmds {
		scmds[i] = cmd.String()
	}
	sort.Strings(scmds)
	return scmds
}

func (c *Broccli) sortedEnv() []string {
	evs := reflect.ValueOf(c.env).MapKeys()
	sevs := make([]string, len(evs))
	for i, ev := range evs {
		sevs[i] = ev.String()
	}
	sort.Strings(sevs)
	return sevs
}

func (c *Broccli) printHelp() {
	fmt.Fprintf(os.Stdout, "%s by %s\n%s\n\n", c.name, c.author, c.usage)
	fmt.Fprintf(os.Stdout, "Usage: %s COMMAND\n\n", path.Base(os.Args[0]))

	if len(c.env) > 0 {
		fmt.Fprintf(os.Stdout, "Required environment variables:\n")
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 8, 8, 0, '\t', 0)
		for _, n := range c.sortedEnv() {
			fmt.Fprintf(w, "%s\t%s\n", n, c.env[n].usage)
		}
		w.Flush()
	}

	fmt.Fprintf(os.Stdout, "Commands:\n")
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 10, 8, 0, '\t', 0)
	for _, n := range c.sortedCommands() {
		fmt.Fprintf(w, "  %s\t%s\n", n, c.commands[n].usage)
	}
	w.Flush()

	fmt.Fprintf(os.Stdout, "\nRun '%s COMMAND --help' for command syntax.\n", path.Base(os.Args[0]))
}

func (c *Broccli) printInvalidCommand(cmd string) {
	fmt.Fprintf(os.Stderr, "Invalid command: %s\n\n", cmd)
	c.printHelp()
}

// getFlagSetPtrs creates flagset instance, parses flags and returns list of pointers to results of parsing the flags.
func (c *Broccli) getFlagSetPtrs(cmd *Command) (map[string]interface{}, map[string]interface{}, []string) {
	fset := flag.NewFlagSet("flagset", flag.ContinueOnError)
	// nothing should come out of flagset
	fset.Usage = func() {}
	fset.SetOutput(io.Discard)

	nameFlags := make(map[string]interface{})
	aliasFlags := make(map[string]interface{})
	fs := cmd.sortedFlags()
	for _, n := range fs {
		f := cmd.flags[n]
		if f.valueType == TypeBool {
			nameFlags[n] = fset.Bool(n, false, "")
			aliasFlags[f.alias] = fset.Bool(f.alias, false, "")
		} else {
			nameFlags[n] = fset.String(n, "", "")
			aliasFlags[f.alias] = fset.String(f.alias, "", "")
		}
	}
	fset.Parse(os.Args[2:])
	return nameFlags, aliasFlags, fset.Args()
}

func (c *Broccli) checkEnv(cmd *Command) int {
	if len(cmd.env) == 0 {
		return 0
	}

	for env, envVar := range cmd.env {
		v := os.Getenv(env)
		envVar.flags = envVar.flags | IsRequired
		err := envVar.validateValue(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s %s: %s\n", c.getParamTypeName(ParamEnvVar), envVar.name, err.Error())
			cmd.printHelp()
			return 1
		}
	}

	return 0
}

func (c *Broccli) processOnTrue(cmd *Command, fs []string, nflags map[string]interface{}, aflags map[string]interface{}) {
	for _, name := range fs {
		if cmd.flags[name].valueType != TypeBool {
			continue
		}

		if cmd.flags[name].options.onTrue == nil {
			continue
		}

		// OnTrue is called when a flag is true
		if *(nflags[name]).(*bool) || *(aflags[cmd.flags[name].alias]).(*bool) {
			cmd.flags[name].options.onTrue(cmd)
		}
	}
}

func (c *Broccli) processFlags(cmd *Command, fs []string, nflags map[string]interface{}, aflags map[string]interface{}) int {
	for _, name := range fs {
		flag := cmd.flags[name]

		if flag.valueType == TypeBool {
			c.parsedFlags[name] = "false"
			if *(nflags[name]).(*bool) || *(aflags[cmd.flags[name].alias]).(*bool) {
				c.parsedFlags[name] = "true"
			}
			continue
		}

		aliasValue := *(aflags[flag.alias]).(*string)
		nameValue := *(nflags[name]).(*string)
		if nameValue != "" && aliasValue != "" {
			fmt.Fprintf(os.Stderr, "ERROR: Both -%s and --%s passed", flag.alias, flag.name)
			return 1
		}
		v := aliasValue
		if nameValue != "" {
			v = nameValue
		}

		err := flag.validateValue(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s %s: %s\n", c.getParamTypeName(ParamFlag), name, err.Error())
			cmd.printHelp()
			return 1
		}

		c.parsedFlags[name] = v
	}

	return 0
}

func (c *Broccli) processArgs(cmd *Command, as []string, args []string) int {
	for i, n := range as {
		v := ""
		if len(args) >= i+1 {
			v = args[i]
		}

		err := cmd.args[n].validateValue(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s %s: %s\n", c.getParamTypeName(ParamArg), cmd.args[n].valuePlaceholder, err.Error())
			cmd.printHelp()
			return 1
		}

		c.parsedArgs[n] = v
	}

	return 0
}

func (c *Broccli) processOnPostValidation(cmd *Command) int {
	if cmd.options.onPostValidation == nil {
		return 0
	}

	err := cmd.options.onPostValidation(cmd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		cmd.printHelp()
		return 1
	}

	return 0
}

func (c *Broccli) parseFlags(cmd *Command) int {
	// check required environment variables
	if exitCode := c.checkEnv(cmd); exitCode != 0 {
		return exitCode
	}

	fs := cmd.sortedFlags()
	nameFlags, aliasFlags, args := c.getFlagSetPtrs(cmd)

	// Loop through boolean flags and execute onTrue() hook if exists.  That function might be used to change behaviour
	// of other flags, eg. when -e is added, another flag or argument might become required (or obsolete).
	// Bool fields will be parsed out in this loop so no reason to process them again in the next one.
	c.processOnTrue(cmd, fs, nameFlags, aliasFlags)

	if exitCode := c.processFlags(cmd, fs, nameFlags, aliasFlags); exitCode != 0 {
		return exitCode
	}

	as := cmd.sortedArgs()
	if exitCode := c.processArgs(cmd, as, args); exitCode != 0 {
		return exitCode
	}

	if exitCode := c.processOnPostValidation(cmd); exitCode != 0 {
		return exitCode
	}

	return 0
}

func (c *Broccli) getParamTypeName(t int8) string {
	if t == ParamArg {
		return "Argument"
	}
	if t == ParamEnvVar {
		return "Env var"
	}
	return "Flag"
}
