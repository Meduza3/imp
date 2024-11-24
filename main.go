package main

import (
	"fmt"
	"os"
	"os/user"

	"github.com/Meduza3/imp/repl"
)

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Witaj %s! To jest imp\n", user.Username)

	// Check if a file is provided as a command-line argument
	if len(os.Args) > 1 {
		file, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		repl.Start(file, os.Stdout) // Use the file as input
	} else {
		repl.Start(os.Stdin, os.Stdout) // Default to standard input
	}
}
