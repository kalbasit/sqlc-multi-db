package main

import (
	"log"
	"os"
	"strings"

	"github.com/kalbasit/sqlc-multi-db/generator"
)

func main() {
	var querierPath string
	// Handle cases where go run might pass "--"
	for _, arg := range os.Args[1:] {
		if arg != "--" && !strings.HasPrefix(arg, "-") {
			querierPath = arg

			break
		}
	}

	if querierPath == "" {
		log.Fatalf("USAGE: %s /path/to/source/querier.go", os.Args[0])
	}

	if _, err := os.Stat(querierPath); err != nil {
		log.Fatalf("stat(%q): %s", querierPath, err)
	}

	generator.Run(querierPath)
}
