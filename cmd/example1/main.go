package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	rand "math/rand/v2"

	"github.com/keenbytes/broccli/v3"
)

func main() {
	cli := broccli.NewCLI("example1", "Example app", "author@example.com")

	printCmd := cli.AddCmd("print", "Prints a hello message", printHandler)

	printCmd.AddArg("first-name", "FIRST_NAME", "First name of the person to welcome", broccli.TypeString, broccli.IsRequired)
	printCmd.AddArg("last-name", "LAST_NAME", "Optional last name", broccli.TypeString, 0)
	
	printCmd.AddFlag("language-file", "l", "PATH_TO_FILE", "File containing 'hello' in many languages", broccli.TypePathFile, broccli.IsRegularFile|broccli.IsExistent|broccli.IsRequired)
	printCmd.AddFlag("alternative", "a", "", "Use alternative welcoming", broccli.TypeBool, 0)

	os.Exit(cli.Run())
}

func printHandler(c *broccli.CLI) int {
	langFile := c.Flag("language-file")
	file, err := os.Open(langFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening file %s: %s", langFile, err.Error())
		return 1
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
			line := scanner.Text()
			if line != "" {
				lines = append(lines, line)
			}
	}
	
	i := rand.IntN(len(lines)-1)
	messageArr := strings.Split(lines[i], ":")
	message := messageArr[0]
	if c.Flag("alternative") == "true" {
		message = messageArr[1]
	}

	firstName := c.Arg("first-name")
	lastName := ""
	if c.Arg("last-name") != "" {
		lastName = fmt.Sprintf(" %s", c.Arg("last-name"))
	}

	fmt.Fprintf(os.Stdout, "%s, %s%s!", message, firstName, lastName)

	return 0
}
