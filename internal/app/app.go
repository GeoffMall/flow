package app

import (
	"fmt"
	"io"
	"os"

	"github.com/GeoffMall/flow/internal/cli"
	"github.com/GeoffMall/flow/internal/format"
	_ "github.com/GeoffMall/flow/internal/format/json" // Register JSON format
	_ "github.com/GeoffMall/flow/internal/format/yaml" // Register YAML format
	"github.com/GeoffMall/flow/internal/operation"
)

func Run() {
	f := cli.ParseFlags()
	// Enable color by default unless --no-color is specified
	f.Color = !f.NoColor

	in, inClose, err := openInput(f.InputFile)
	if err != nil {
		fatalf("Error opening input: %v\n", err)
	}
	defer inClose()

	out, outClose, err := openOutput(f.OutputFile)
	if err != nil {
		fatalf("Error opening output: %v\n", err)
	}
	defer outClose()

	if err := run(in, out, f); err != nil {
		fatalf("Processing error: %v\n", err)
	}
}

func openInput(path string) (io.Reader, func(), error) {
	if path == "" {
		return os.Stdin, func() {}, nil
	}
	// #nosec G304 - CLI tool trusts user-provided file paths
	f, err := os.Open(path)
	if err != nil {
		return nil, func() {}, err
	}
	return f, func() { _ = f.Close() }, nil
}

func openOutput(path string) (io.Writer, func(), error) {
	if path == "" {
		return os.Stdout, func() {}, nil
	}
	// #nosec G304 - CLI tool trusts user-provided file paths
	f, err := os.Create(path)
	if err != nil {
		return nil, func() {}, err
	}
	return f, func() { _ = f.Close() }, nil
}

func buildPipeline(opts *cli.Flags) (*operation.Pipeline, error) {
	var ops []operation.Operation

	if len(opts.PickPaths) > 0 {
		ops = append(ops, operation.NewPick(opts.PickPaths, opts.PreserveHierarchy))
	}

	if len(opts.SetPairs) > 0 {
		setOp, err := operation.NewSetFromPairs(opts.SetPairs)
		if err != nil {
			return nil, err
		}
		ops = append(ops, setOp)
	}

	if len(opts.DeletePaths) > 0 {
		ops = append(ops, operation.NewDelete(opts.DeletePaths))
	}

	return operation.NewPipeline(ops...), nil
}

// run executes one full pass: parse stream -> apply pipeline -> print.
func run(in io.Reader, out io.Writer, opts *cli.Flags) error {
	// Build operation pipeline
	pipe, err := buildPipeline(opts)
	if err != nil {
		return err
	}

	// Auto-detect input format
	inputFormat, br, err := format.AutoDetect(in)
	if err != nil {
		return fmt.Errorf("format detection failed: %w", err)
	}

	// Create parser for detected format
	parser, err := inputFormat.NewParser(br)
	if err != nil {
		return fmt.Errorf("failed to create parser: %w", err)
	}

	// Determine output format (default to json if not specified)
	outputFormatName := opts.ToFormat
	if outputFormatName == "" {
		outputFormatName = "json"
	}

	// Get output format
	outputFormat, err := format.Get(outputFormatName)
	if err != nil {
		return fmt.Errorf("unknown output format %q: %w", outputFormatName, err)
	}

	// Create formatter for output
	formatter := outputFormat.NewFormatter(out, format.FormatterOptions{
		Color:   opts.Color,
		Compact: opts.Compact,
	})
	defer formatter.Close()

	// Process stream: parse -> transform -> format
	return parser.ForEach(func(doc any) error {
		outDoc := doc
		if !pipe.Empty() {
			var err error
			outDoc, err = pipe.Apply(doc)
			if err != nil {
				return err
			}
		}
		return formatter.Write(outDoc)
	})
}

func fatalf(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
