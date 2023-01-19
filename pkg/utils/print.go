package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Writer struct {
	out io.Writer
}

// NewWriter returns a new Writer instance. If out is nil (which it can't really
// be unless explicitly set so), it will default to os.Stdout.
func NewWriter(out io.Writer) Writer {
	if out == nil {
		out = os.Stdout
	}

	return Writer{out: out}
}

// Print will write the given string to the Writer's out.
func (w Writer) Print(output interface{}) {
	fmt.Fprintln(w.out, output)
}

// Printf will write the given string(s) to the Writer's out and apply given formatting.
func (w Writer) Printf(format string, output ...interface{}) {
	fmt.Fprintf(w.out, format, output...)
}

// PrettyPrint will write the given object the out writer in nice JSON.
func (w Writer) PrettyPrint(response interface{}) error {
	resJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return err
	}

	fmt.Fprintf(w.out, "%s\n", string(resJSON))

	return nil
}
