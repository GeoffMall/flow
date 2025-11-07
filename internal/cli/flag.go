package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/GeoffMall/flow/internal/version"
)

// ANSI color codes for ASCII art
const (
	cyan1 = "\x1b[38;5;51m" // bright cyan
	cyan2 = "\x1b[38;5;50m" // cyan
	cyan3 = "\x1b[38;5;45m" // aqua cyan
	aqua  = "\x1b[38;5;87m" // light aqua
	reset = "\x1b[0m"       // reset color
)

// Flags holds all parsed command-line arguments.
type Flags struct {
	InputFile         string   // file to read from (optional; defaults to stdin)
	OutputFile        string   // file to write to (optional; defaults to stdout)
	PickPaths         []string // list of dotted paths to pick
	SetPairs          []string // raw key=value strings for --set
	DeletePaths       []string // list of paths to delete
	Color             bool     // pretty colorized output (internal use)
	NoColor           bool     // disable colorized output
	Compact           bool     // minified output
	FromFormat        string   // input format: json | yaml (defaults to json, or auto-detected from file extension)
	ToFormat          string   // convert output format: json | yaml
	PreserveHierarchy bool     // preserve full path structure in pick output (legacy behavior)
	ShowHelp          bool     // show help and exit
	ShowVersion       bool     // show version and exit
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
	flag.BoolVar(&f.NoColor, "no-color", false, "Disable colorized output")
	flag.BoolVar(&f.Compact, "compact", false, "Minify output instead of pretty-printing")
	flag.StringVar(&f.FromFormat, "from", "", "Input format: json | yaml (if not specified, detected from file extension or defaults to json)")
	flag.StringVar(&f.ToFormat, "to", "", "Convert output format: json | yaml")
	flag.BoolVar(&f.PreserveHierarchy, "preserve-hierarchy", false, "Preserve full path structure in pick output (default: false, outputs values like jq)")
	flag.BoolVar(&f.ShowHelp, "help", false, "Show usage")
	flag.BoolVar(&f.ShowVersion, "version", false, "Show version information")

	flag.Usage = usage

	flag.Parse()

	// If help was requested, print and exit
	if f.ShowHelp {
		flag.Usage()
		os.Exit(0)
	}

	// If the version was requested, print and exit
	if f.ShowVersion {
		printVersion()
		os.Exit(0)
	}

	f.PickPaths = pickPaths
	f.SetPairs = setPairs
	f.DeletePaths = deletePaths

	// Validate format flags
	if f.FromFormat != "" && f.FromFormat != "json" && f.FromFormat != "yaml" {
		printLinef("Error: invalid format '%s' for --from flag. Supported formats are 'json' and 'yaml'.\n", f.FromFormat)
		flag.Usage()
		os.Exit(1)
	}

	if f.ToFormat != "" && f.ToFormat != "json" && f.ToFormat != "yaml" {
		printLinef("Error: invalid format '%s' for --to flag. Supported formats are 'json' and 'yaml'.\n", f.ToFormat)
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

// asciiArt returns the colored ASCII art banner for "flow"
func asciiArt() string {
	art := cyan1 + "######## ##        " + cyan2 + "#######  " + cyan3 + "##      ## " + reset + "\n"
	art += cyan1 + "##       ##       " + cyan2 + "##     ## " + cyan3 + "##  ##  ## " + reset + "\n"
	art += cyan1 + "##       ##       " + cyan2 + "##     ## " + cyan3 + "##  ##  ## " + reset + "\n"
	art += cyan2 + "######   ##       " + cyan3 + "##     ## " + aqua + "##  ##  ## " + reset + "\n"
	art += cyan2 + "##       ##       " + cyan3 + "##     ## " + aqua + "##  ##  ## " + reset + "\n"
	art += cyan2 + "##       ##       " + cyan3 + "##     ## " + aqua + "##  ##  ## " + reset + "\n"
	art += cyan3 + "##       ########  " + aqua + "#######   ###  ###  " + reset + "\n"
	art += "\n" + aqua + "         ~stream your data~" + reset + "\n\n"
	return art
}

func usage() {
	// Display ASCII art banner at the top
	printLinef("%s", asciiArt())
	printLinef("Usage: flow [flags]\n\n")
	printLinef("Examples:\n")
	printLinef("  cat data.json | flow --pick user.name --pick user.id  # outputs: {\"name\": \"alice\", \"id\": 7}\n")
	printLinef("  cat data.json | flow --pick user.name                 # outputs: \"alice\"\n")
	printLinef("  flow config.yaml --set server.port=8080 --delete debug --to json\n")
	printLinef("\nFlags:\n")
	flag.PrintDefaults()
}

func printLinef(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

// printVersion prints the version information
func printVersion() {
	info := version.Get()
	fmt.Println(info.String())
}
