package main

import (
	"fmt"
	"os"
	"runa/internal/menu"
)

func main() {
	if err := menu.Run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "red:", err)
		os.Exit(1)
	}
}
