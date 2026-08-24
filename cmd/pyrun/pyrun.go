package main

import (
	"fmt"
	"os"

	"github.com/asiangoldfish/python-runner/lib"
)

const VERSION = "1.2.4"

var VERBOSE = false

func main() {
	// Help pages
	if len(os.Args) < 2 {
		lib.Usage()
		return
	}

	// We mark the starting index for options, as we parse the flags first.
	startIndexForOptions := -1

	// Find flags. These alter the program's behaviour.
	for i, flag := range os.Args[1:] {
		if flag[0] != '-' {
			break
		}

		switch flag {
		case "-v":
			VERBOSE = true
		default:
			fmt.Fprintln(os.Stderr, "Flag "+flag+" is not recognised.")
			os.Exit(1)
		}

		startIndexForOptions = i
	}

	rest := os.Args[startIndexForOptions+2:]

	// No option was provided.
	if startIndexForOptions <= 1 && len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "Flags were provided, but options are missing.")
		return
	}

	// Parse arguments
	switch rest[0] {
	case "cmd":
		var cmdArgs []string
		var cmdName string
		if len(rest) >= 2 {
			cmdName = rest[1]
		}
		if len(rest) >= 3 {
			cmdArgs = rest[2:]
		}
		if err := lib.ExecuteCmd(cmdName, cmdArgs, "."); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	case "init":
		if !lib.Initialise(VERBOSE) {
			os.Exit(1)
		}
	case "-h", "--help", "help":
		lib.Usage()
	case "install":
		if err := lib.InstallPackage(rest[1:], VERBOSE, "."); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "run":
		if err := lib.Run(rest[1:]); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	case "version":
		fmt.Println("pyrun " + VERSION)
	default:
		fmt.Fprintln(os.Stderr, "Option '"+rest[0]+"' is not recognised.")
	}
}
