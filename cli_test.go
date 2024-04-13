package broccli

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"
)

// TestCLI creates a test CLI instance with a single command with flags and test
// basic functionality.
func TestCLI(t *testing.T) {
	f, err := os.CreateTemp("", "stdout")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(f.Name())

	devNull, err := os.OpenFile("/dev/null", os.O_APPEND, 644)
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout = devNull
	os.Stderr = devNull

	c := NewCLI("Example", "App", "Author <a@example.com>")
	cmd1 := c.AddCmd("cmd1", "Prints out a string", func(c *CLI) int {
		fmt.Fprintf(f, "TESTVALUE:%s%s\n\n", c.Flag("tekst"), c.Flag("alphanumdots"))
		return 2
	})
	cmd1.AddFlag("tekst", "t", "Text", "Text to print", TypeString, IsRequired)
	cmd1.AddFlag("alphanumdots", "a", "Alphanum with dots", "Can have dots", TypeAlphanumeric, AllowDots)
	cmd1.AddFlag("make-required", "r", "", "Make alphanumdots required", TypeBool, 0, OnTrue(func(c *Cmd) {
		c.flags["alphanumdots"].flags = c.flags["alphanumdots"].flags | IsRequired
	}))

	os.Args = []string{"test", "cmd1"}
	got := c.Run()
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "-t", ""}
	got = c.Run()
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "--tekst", "Tekst123", "--alphanumdots"}
	got = c.Run()
	if got != 2 {
		t.Errorf("CLI.Run() should have returned 2 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "--tekst", "Tekst123", "-r"}
	got = c.Run()
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "--tekst", "Tekst123", "--alphanumdots", "aZ0-9"}
	got = c.Run()
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "--tekst", "Tekst123", "--alphanumdots", "aZ0.9"}
	got = c.Run()
	if got != 2 {
		t.Errorf("CLI.Run() should have returned 2 instead of %d", got)
	}

	f2, err := os.Open(f.Name())
	if err != nil {
		log.Fatal(err)
	}
	defer f2.Close()
	b, err := io.ReadAll(f2)
	if err != nil {
		log.Fatal(err)
	}

	if !strings.Contains(string(b), "TESTVALUE:Tekst123aZ0.9") {
		t.Errorf("Cmd handler failed to work")
	}
}
