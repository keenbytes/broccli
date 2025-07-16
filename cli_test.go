package broccli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

// TestCLI creates a test CLI instance with a single command with flags and test
// basic functionality.
func TestCLI(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Error("error creating temporary file")
	}

	defer func() {
		err := os.Remove(f.Name())
		if err != nil {
			t.Error("error removing temporary file")
		}
	}()

	devNull, err := os.OpenFile("/dev/null", os.O_APPEND, 0o600)
	if err != nil {
		t.Error("error opening temporary file")
	}

	os.Stdout = devNull
	os.Stderr = devNull

	c := NewBroccli("Example", "App", "Author <a@example.com>")
	cmd1 := c.Command("cmd1", "Prints out a string", func(_ context.Context, c *Broccli) int {
		_, _ = fmt.Fprintf(f, "TESTVALUE:%s%s\n\n", c.Flag("tekst"), c.Flag("alphanumdots"))

		if c.Flag("bool") == "true" {
			_, _ = fmt.Fprintf(f, "BOOL:true")
		}

		return 2
	})
	cmd1.Flag("tekst", "t", "Text", "Text to print", TypeString, IsRequired)
	cmd1.Flag(
		"alphanumdots",
		"a",
		"Alphanum with dots",
		"Can have dots",
		TypeAlphanumeric,
		AllowDots,
	)
	cmd1.Flag(
		"make-required",
		"r",
		"",
		"Make alphanumdots required",
		TypeBool,
		0,
		OnTrue(func(c *Command) {
			c.flags["alphanumdots"].flags |= IsRequired
		}),
	)
	// Boolean should work fine even when the optional OnTrue is not passed
	cmd1.Flag("bool", "b", "", "Bool value", TypeBool, 0)

	os.Args = []string{"test", "cmd1"}

	got := c.Run(context.Background())
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "-t", ""}

	got = c.Run(context.Background())
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "--tekst", "Tekst123", "--alphanumdots"}

	got = c.Run(context.Background())
	if got != 2 {
		t.Errorf("CLI.Run() should have returned 2 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "--tekst", "Tekst123", "-r"}

	got = c.Run(context.Background())
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "--tekst", "Tekst123", "--alphanumdots", "aZ0-9"}

	got = c.Run(context.Background())
	if got != 1 {
		t.Errorf("CLI.Run() should have returned 1 instead of %d", got)
	}

	os.Args = []string{"test", "cmd1", "--tekst", "Tekst123", "--alphanumdots", "aZ0.9", "-b"}

	got = c.Run(context.Background())
	if got != 2 {
		t.Errorf("CLI.Run() should have returned 2 instead of %d", got)
	}

	f2, err := os.Open(f.Name())
	if err != nil {
		t.Error("error opening temporary file")
	}

	defer func() {
		err := f2.Close()
		if err != nil {
			t.Error("error closing temporary file")
		}
	}()

	b, err := io.ReadAll(f2)
	if err != nil {
		t.Error("error reading output file contents")
	}

	if !strings.Contains(string(b), "TESTVALUE:Tekst123aZ0.9") {
		t.Errorf("Cmd handler failed to work")
	}

	if !strings.Contains(string(b), "BOOL:true") {
		t.Errorf("Cmd handler failed to work")
	}
}
