package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kalbasit/sqlc-multi-db/generator"
)

var errInvalidEngineFormat = errors.New("invalid engine format: expected name:package")

type engineFlag []generator.Engine

func (e *engineFlag) String() string {
	parts := make([]string, len(*e))
	for i, eng := range *e {
		parts[i] = eng.Name + ":" + eng.Package
	}

	return strings.Join(parts, ", ")
}

func (e *engineFlag) Set(value string) error {
	parts := strings.SplitN(value, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("%w: got %q", errInvalidEngineFormat, value)
	}

	*e = append(*e, generator.Engine{Name: parts[0], Package: parts[1]})

	return nil
}

func main() {
	var engines engineFlag

	flag.Var(&engines, "engine", "Engine in name:package format (repeatable)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "USAGE: %s [--engine name:package ...] /path/to/source/querier.go\n", os.Args[0])
	}

	flag.Parse()

	if len(engines) == 0 {
		log.Fatalf("USAGE: sqlc-multi-db --engine name:package [--engine ...] /path/to/querier.go")
	}

	args := flag.Args()
	if len(args) == 0 {
		log.Fatalf("USAGE: sqlc-multi-db --engine name:package [--engine ...] /path/to/querier.go")
	}

	querierPath := args[0]

	if _, err := os.Stat(querierPath); err != nil {
		log.Fatalf("stat(%q): %s", querierPath, err)
	}

	generator.Run(querierPath, []generator.Engine(engines))
}
