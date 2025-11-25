package main

import (
	"fmt"
	"os"

	"github.com/donbowman/github-multi-account-manager/internal/tui"
)

func main() {
	if err := tui.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
