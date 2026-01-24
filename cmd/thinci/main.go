package main

import (
	"fmt"
	"os"

	"github.com/sourceplane/sourceplane/cmd"
)

func main() {
	// Set program name for thinci
	os.Args[0] = "thinci"

	if err := cmd.ExecuteThinCI(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
