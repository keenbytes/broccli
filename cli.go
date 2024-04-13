package broccli

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"sort"
	"text/tabwriter"
)

// CLI is main CLI application definition.
// It has a name, description, author which are printed out to the screen in the usage syntax.
// Each CLI have commands (represented by Cmd).  Optionally, it is possible to require environment
// variables.
type CLI struct {
	name        string
	desc        string
	author      string
	cmds        map[string]*Cmd
	envVars     map[string]*param
	parsedFlags map[string]string
	parsedArgs  map[string]string
}

// NewCLI returns pointer to a new CLI instance with specified name, description and author.  All these are used when
// displaying syntax help.
func NewCLI(n string, d string, a string) *CLI {
	c := &CLI{
		name:        n,
		desc:        d,
		author:      a,
		cmds:        map[string]*Cmd{},
		envVars:     map[string]*param{},
		parsedFlags: map[string]string{},
		parsedArgs:  map[string]string{},
	}
	return c
}

// AddCmd returns pointer to a new command with specified name, description and handler.  Handler is a function that
// gets called when command is executed.
// Additionally, there is a set of options that can be passed as arguments.  Search for cmdOption for more info.
func (c *CLI) AddCmd(n string, d string, h func(cli *CLI) int, opts ...cmdOption) *Cmd {
	c.cmds[n] = &Cmd{
		name:    n,
		desc:    d,
		flags:   map[string]*param{},
		args:    map[string]*param{},
		envVars: map[string]*param{},
		handler: h,
		options: cmdOptions{},
	}
	for _, o := range opts {
		o(&(c.cmds[n].options))
	}
	return c.cmds[n]
}

// AddEnvVar returns pointer to a new environment variable that is required to run every command.
// Method requires name, eg. MY_VAR, and description.
func (c *CLI) AddEnvVar(n string, d string) {
	c.envVars[n] = &param{
		name:    n,
		desc:    d,
		flags:   IsRequired,
		options: paramOptions{},
	}
}

// Flag returns value of flag.
func (c *CLI) Flag(n string) string {
	return c.parsedFlags[n]
}

// Arg returns value of arg.
func (c *CLI) Arg(n string) string {
	return c.parsedArgs[n]
}

// Run parses the arguments, validates them and executes command handler.
// In case of invalid arguments, error is printed to stderr and 1 is returned.  Return value should be treated as exit
// code.
func (c *CLI) Run() int {
	// display help, first arg is binary filename
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		c.printHelp()
		return 0
	}
	for _, n := range c.sortedCmds() {
		if n != os.Args[1] {
			continue
		}
		// display command help
		if len(os.Args) > 2 && (os.Args[2] == "-h" || os.Args[2] == "--help") {
			c.cmds[n].printHelp()
			return 0
		}

		// check required environment variables
		if len(c.envVars) > 0 {
			for env, param := range c.envVars {
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
		exitCode := c.parseFlags(c.cmds[n])
		if exitCode > 0 {
			return exitCode
		}

		return c.cmds[n].handler(c)
	}

	// command not found
	c.printInvalidCmd(os.Args[1])
	return 1
}

func (c *CLI) sortedCmds() []string {
	cmds := reflect.ValueOf(c.cmds).MapKeys()
	scmds := make([]string, len(cmds))
	for i, cmd := range cmds {
		scmds[i] = cmd.String()
	}
	sort.Strings(scmds)
	return scmds
}

func (c *CLI) sortedEnvVars() []string {
	evs := reflect.ValueOf(c.envVars).MapKeys()
	sevs := make([]string, len(evs))
	for i, ev := range evs {
		sevs[i] = ev.String()
	}
	sort.Strings(sevs)
	return sevs
}

func (c *CLI) printHelp() {
	fmt.Fprintf(os.Stdout, "%s by %s\n%s\n\n", c.name, c.author, c.desc)
	fmt.Fprintf(os.Stdout, "Usage: %s COMMAND\n\n", path.Base(os.Args[0]))

	if len(c.envVars) > 0 {
		fmt.Fprintf(os.Stdout, "Required environment variables:\n")
		w := new(tabwriter.Writer)
		w.Init(os.Stdout, 8, 8, 0, '\t', 0)
		for _, n := range c.sortedEnvVars() {
			fmt.Fprintf(w, "%s\t%s\n", n, c.envVars[n].desc)
		}
		w.Flush()
	}

	fmt.Fprintf(os.Stdout, "Commands:\n")
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 10, 8, 0, '\t', 0)
	for _, n := range c.sortedCmds() {
		fmt.Fprintf(w, "  %s\t%s\n", n, c.cmds[n].desc)
	}
	w.Flush()

	fmt.Fprintf(os.Stdout, "\nRun '%s COMMAND --help' for command syntax.\n", path.Base(os.Args[0]))
}

func (c *CLI) printInvalidCmd(cmd string) {
	fmt.Fprintf(os.Stderr, "Invalid command: %s\n\n", cmd)
	c.printHelp()
}

// getFlagSetPtrs creates flagset instance, parses flags and returns list of pointers to results of parsing the flags.
func (c *CLI) getFlagSetPtrs(cmd *Cmd) (map[string]interface{}, map[string]interface{}, []string) {
	fset := flag.NewFlagSet("flagset", flag.ContinueOnError)
	// nothing should come out of flagset
	fset.Usage = func() {}
	fset.SetOutput(ioutil.Discard)

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

func (c *CLI) checkEnvVars(cmd *Cmd) int {
	if len(cmd.envVars) == 0 {
		return 0
	}

	for env, envVar := range cmd.envVars {
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

func (c *CLI) processOnTrue(cmd *Cmd, fs []string, nflags map[string]interface{}, aflags map[string]interface{}) {
	for _, name := range fs {
		if cmd.flags[name].valueType != TypeBool {
			continue
		}

		if cmd.flags[name].options.onTrue == nil {
			continue
		}

		c.parsedFlags[name] = "false"
		if *(nflags[name]).(*bool) == true || *(aflags[cmd.flags[name].alias]).(*bool) == true {
			c.parsedFlags[name] = "true"
			cmd.flags[name].options.onTrue(cmd)
		}
	}
}

func (c *CLI) processFlags(cmd *Cmd, fs []string, nflags map[string]interface{}, aflags map[string]interface{}) int {
	for _, name := range fs {
		flag := cmd.flags[name]

		if flag.valueType == TypeBool {
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

func (c *CLI) processArgs(cmd *Cmd, as []string, args []string) int {
	for i, n := range as {
		v := ""
		if len(args) >= i+1 {
			v = args[i]
		}

		err := cmd.args[n].validateValue(v)
		if err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %s %s: %s\n", c.getParamTypeName(ParamArg), cmd.args[n].helpValue, err.Error())
			cmd.printHelp()
			return 1
		}

		c.parsedArgs[n] = v
	}

	return 0
}

func (c *CLI) processOnPostValidation(cmd *Cmd) int {
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

func (c *CLI) parseFlags(cmd *Cmd) int {
	// check required environment variables
	if exitCode := c.checkEnvVars(cmd); exitCode != 0 {
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

func (c *CLI) getParamTypeName(t int8) string {
	if t == ParamArg {
		return "Argument"
	}
	if t == ParamEnvVar {
		return "Env var"
	}
	return "Flag"
}
