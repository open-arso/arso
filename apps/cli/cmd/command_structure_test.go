package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestRootCommandContainsExpectedSubcommands(t *testing.T) {
	paths := [][]string{
		{"config"},
		{"config", "init"},
		{"config", "path"},
		{"config", "show"},
		{"config", "get"},
		{"config", "set"},
		{"find"},
		{"node"},
		{"pass"},
		{"pass", "list"},
		{"pass", "next"},
		{"version"},
		{"weather"},
	}

	for _, path := range paths {
		command, _, err := rootCmd.Find(path)
		if err != nil {
			t.Fatalf("Find(%v) returned error: %v", path, err)
		}

		if command == nil {
			t.Fatalf("Find(%v) returned nil command", path)
		}

		if got, want := command.Name(), path[len(path)-1]; got != want {
			t.Fatalf("Find(%v) command = %q, want %q", path, got, want)
		}
	}
}

func TestCommandFlagDefaults(t *testing.T) {
	tests := []struct {
		name    string
		command *cobra.Command
		flag    string
		want    string
	}{
		{name: "root config flag", command: rootCmd, flag: "config", want: ""},
		{name: "find output", command: findCmd, flag: "output", want: "text"},
		{name: "find at", command: findCmd, flag: "at", want: ""},
		{name: "find elements", command: findCmd, flag: "elements", want: "false"},
		{name: "list output", command: listCmd, flag: "output", want: "text"},
		{name: "list from", command: listCmd, flag: "from", want: ""},
		{name: "list lookahead", command: listCmd, flag: "lookahead", want: "48h"},
		{name: "list min elevation", command: listCmd, flag: "min-elevation", want: "10"},
		{name: "next output", command: nextCmd, flag: "output", want: "text"},
		{name: "next from", command: nextCmd, flag: "from", want: ""},
		{name: "next lookahead", command: nextCmd, flag: "lookahead", want: "48h"},
		{name: "next min elevation", command: nextCmd, flag: "min-elevation", want: "10"},
		{name: "version output", command: versionCmd, flag: "output", want: "text"},
		{name: "config overwrite", command: configInitCmd, flag: "overwrite", want: "false"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag := tt.command.Flags().Lookup(tt.flag)
			if flag == nil {
				t.Fatalf("Lookup(%q) returned nil", tt.flag)
			}

			if got := flag.DefValue; got != tt.want {
				t.Fatalf("flag %q default = %q, want %q", tt.flag, got, tt.want)
			}
		})
	}
}

func TestCommandArgumentValidation(t *testing.T) {
	tests := []struct {
		name    string
		command *cobra.Command
		args    []string
		wantErr bool
	}{
		{name: "find accepts one arg", command: findCmd, args: []string{"ISS"}},
		{name: "find rejects zero args", command: findCmd, wantErr: true},
		{name: "list accepts one arg", command: listCmd, args: []string{"ISS"}},
		{name: "list rejects two args", command: listCmd, args: []string{"ISS", "HUBBLE"}, wantErr: true},
		{name: "next accepts one arg", command: nextCmd, args: []string{"ISS"}},
		{name: "next rejects zero args", command: nextCmd, wantErr: true},
		{name: "config get accepts one arg", command: configGetCmd, args: []string{"node.name"}},
		{name: "config get rejects zero args", command: configGetCmd, wantErr: true},
		{name: "config set accepts two args", command: configSetCmd, args: []string{"node.name", "paris"}},
		{name: "config set rejects one arg", command: configSetCmd, args: []string{"node.name"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.command.Args(tt.command, tt.args)

			if tt.wantErr && err == nil {
				t.Fatal("Args() expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("Args() unexpected error: %v", err)
			}
		})
	}
}

func TestFlagParsingUpdatesBoundVariables(t *testing.T) {
	originalListOutput := outputList
	originalListLookahead := lookaheadList
	originalListMinElevation := minElevationList
	originalNextOutput := output
	originalNextLookahead := lookahead
	originalFindElements := findElements
	originalVersionOutput := versionOutput

	t.Cleanup(func() {
		outputList = originalListOutput
		lookaheadList = originalListLookahead
		minElevationList = originalListMinElevation
		output = originalNextOutput
		lookahead = originalNextLookahead
		findElements = originalFindElements
		versionOutput = originalVersionOutput
	})

	if err := listCmd.Flags().Set("output", "json"); err != nil {
		t.Fatalf("list output Set() error = %v", err)
	}
	if err := listCmd.Flags().Set("lookahead", "72h"); err != nil {
		t.Fatalf("list lookahead Set() error = %v", err)
	}
	if err := listCmd.Flags().Set("min-elevation", "25"); err != nil {
		t.Fatalf("list min-elevation Set() error = %v", err)
	}
	if err := nextCmd.Flags().Set("output", "ndjson"); err != nil {
		t.Fatalf("next output Set() error = %v", err)
	}
	if err := nextCmd.Flags().Set("lookahead", "6h"); err != nil {
		t.Fatalf("next lookahead Set() error = %v", err)
	}
	if err := findCmd.Flags().Set("elements", "true"); err != nil {
		t.Fatalf("find elements Set() error = %v", err)
	}
	if err := versionCmd.Flags().Set("output", "json"); err != nil {
		t.Fatalf("version output Set() error = %v", err)
	}

	if outputList != "json" {
		t.Fatalf("outputList = %q, want %q", outputList, "json")
	}
	if lookaheadList != "72h" {
		t.Fatalf("lookaheadList = %q, want %q", lookaheadList, "72h")
	}
	if minElevationList != 25 {
		t.Fatalf("minElevationList = %d, want %d", minElevationList, 25)
	}
	if output != "ndjson" {
		t.Fatalf("output = %q, want %q", output, "ndjson")
	}
	if lookahead != "6h" {
		t.Fatalf("lookahead = %q, want %q", lookahead, "6h")
	}
	if !findElements {
		t.Fatal("findElements = false, want true")
	}
	if versionOutput != "json" {
		t.Fatalf("versionOutput = %q, want %q", versionOutput, "json")
	}
}
