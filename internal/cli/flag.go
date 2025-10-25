package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// Flags holds all parsed command-line arguments.
type Flags struct {
	InputFile   string   // file to read from (optional; defaults to stdin)
	OutputFile  string   // file to write to (optional; defaults to stdout)
	PickPaths   []string // list of dotted paths to pick
	SetPairs    []string // raw key=value strings for --set
	DeletePaths []string // list of paths to delete
	Color       bool     // pretty colorized output
	Compact     bool     // minified output
	ToFormat    string   // convert output format: json | yaml
	ShowHelp    bool
}

// ParseFlags parses CLI flags and returns a populated Flags struct.
// It exits with a usage message if invalid flags are provided.
func ParseFlags() *Flags {
	f := &Flags{}

	// Define repeatable flags by creating custom flag slices
	var pickPaths multiStringFlag
	var setPairs multiStringFlag
	var deletePaths multiStringFlag

	flag.Var(&pickPaths, "pick", "Pick a key or path from the input (can be used multiple times)")
	flag.Var(&setPairs, "set", "Set a key to a value (format: path=value, can be used multiple times)")
	flag.Var(&deletePaths, "delete", "Delete a key or path from the input (can be used multiple times)")

	flag.StringVar(&f.InputFile, "in", "", "Path to input file (optional, defaults to stdin)")
	flag.StringVar(&f.OutputFile, "out", "", "Path to output file (optional, defaults to stdout)")
	flag.BoolVar(&f.Color, "color", false, "Enable colorized output")
	flag.BoolVar(&f.Compact, "compact", false, "Minify output instead of pretty-printing")
	flag.StringVar(&f.ToFormat, "to", "", "Convert output format: json | yaml")
	flag.BoolVar(&f.ShowHelp, "help", false, "Show usage")

	flag.Usage = usage

	flag.Parse()

	// If help was requested, print and exit
	if f.ShowHelp {
		flag.Usage()
		os.Exit(0)
	}

	f.PickPaths = pickPaths
	f.SetPairs = setPairs
	f.DeletePaths = deletePaths

	// Additional validation can be added here if needed
	if f.ToFormat != "" && f.ToFormat != "json" && f.ToFormat != "yaml" {
		printLine("Error: invalid format '%s' for --to flag. Supported formats are 'json' and 'yaml'.\n", f.ToFormat)
		flag.Usage()
		os.Exit(1)
	}

	return f
}

type multiStringFlag []string

func (m *multiStringFlag) String() string {
	return strings.Join(*m, ", ")
}

func (m *multiStringFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

func usage() {
	// Custom usage message can be defined here
	printLine("Usage: flow [flags]\n\n")
	printLine("Examples:\n")
	printLine("  cat data.json | flow --pick user.name --pick user.id\n")
	printLine("  flow config.yaml --set server.port=8080 --delete debug --to json\n")
	printLine("\nFlags:\n")
	flag.PrintDefaults()
}

func printLine(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}
