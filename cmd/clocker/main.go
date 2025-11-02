package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sporadisk/clocker/calculator"
	"github.com/sporadisk/clocker/client/logfile"
	"github.com/sporadisk/clocker/config"
)

const helpMsg = `
Valid flags:
  --file
    Path to a file on the local FS, which will be watched for changes and used as input.

`

func main() {
	validInput, err := run()
	if err != nil {
		if !validInput {
			fmt.Print(helpMsg)
		}

		fmt.Printf("Error: %s\n", err.Error())

		os.Exit(1)
		return
	}
}

func run() (validInput bool, err error) {
	confPath := flag.String("config", "", "Path to config file")
	filePath := flag.String("file", "", "Path to a local file to watch for changes")
	flag.Parse()

	if *confPath != "" {
		fmt.Printf("Using config file: %s\n", *confPath)
	}

	conf, err := config.Load(*confPath)
	if err != nil {
		return false, fmt.Errorf("config.Load: %w", err)
	}

	if *filePath == "" {
		return false, fmt.Errorf("--file argument is missing.")
	}

	subscriber, err := logfile.NewSubscriber(*filePath)
	if err != nil {
		return true, fmt.Errorf("logfile.NewSubscriber: %w", err)
	}

	calc := &calculator.Calculator{
		Conf:       conf,
		Subscriber: subscriber,
	}

	err = calc.Start()
	if err != nil {
		return true, fmt.Errorf("app.Start: %w", err)
	}

	return true, nil
}
