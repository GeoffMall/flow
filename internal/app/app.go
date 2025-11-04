package app

import (
	"fmt"
	"io"
	"os"

	"github.com/GeoffMall/flow/internal/cli"
	"github.com/GeoffMall/flow/internal/operation"
	"github.com/GeoffMall/flow/internal/parser"
	"github.com/GeoffMall/flow/internal/printer"
)

func Run() {
	f := cli.ParseFlags()
	// default to color
	f.Color = true

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
	p := printer.New(printer.Options{
		ToFormat: opts.ToFormat,
		Color:    opts.Color,
		Compact:  opts.Compact,
		Writer:   out,
	})

	pipe, err := buildPipeline(opts)
	if err != nil {
		return err
	}

	pr, err := parser.New(in)
	if err != nil {
		return err
	}

	return pr.ForEach(func(doc any) error {
		outDoc := doc
		if !pipe.Empty() {
			var err error
			outDoc, err = pipe.Apply(doc)
			if err != nil {
				return err
			}
		}
		return p.Write(outDoc)
	})
}

func fatalf(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
	os.Exit(1)
}
