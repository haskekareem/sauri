package main

import (
	"errors"
	"github.com/fatih/color"
	"github.com/haskekareem/sauri"
	"os"
)

const version = "1.0.0"

var sauri2 sauri.Sauri

// Main entry point for the command line tool
func main() {
	var message string

	// arg 1 = ./sauri: load the command line arguments
	arg2, arg3, arg4, err := validateInputs()
	if err != nil {
		exitGracefully(errors.New("you must select any of the above commands"))
	}

	// load setup
	setUp(arg2)

	switch arg2 {
	case "help":
		showHelp()
	case "new":
		if arg3 == "" {
			exitGracefully(errors.New("new require an application name"))
		}
		doNew(arg3)
	case "version":
		color.Yellow("Application version: " + version)
	case "make":
		if arg3 == "" {
			exitGracefully(errors.New("make required a subcommand: (migration|handlers)"))
		}
		err = doMake(arg3, arg4)
		if err != nil {
			exitGracefully(err)
		}
	case "migrate":
		//push the migration files to the database
		// migrate up as the default setting
		if arg3 == "" {
			arg3 = "up"
		}
		err = doMigrate(arg3, arg4)
		if err != nil {
			exitGracefully(err)
		}
		message = "migrations complete!"
	default:
		showHelp()
	}
	exitGracefully(nil, message)
}

func validateInputs() (string, string, string, error) {
	var arg2, arg3, arg4 string

	if len(os.Args) > 1 {
		arg2 = os.Args[1]

		if len(os.Args) >= 3 {
			arg3 = os.Args[2]
		}
		if len(os.Args) >= 4 {
			arg4 = os.Args[3]
		}
	} else {
		// first argument in the command line
		color.Red("Error: command required")
		showHelp()
		return "", "", "", errors.New("command required")

	}
	return arg2, arg3, arg4, nil
}
