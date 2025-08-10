package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Printf("usage: %s COMMAND [options]\n", os.Args[0])
	fmt.Println("")
	fmt.Println("commands")
	fmt.Println("serve - serve the data in the sqlite db over http")
	fmt.Println("load - load data from htg")
	fmt.Println("")
	fmt.Println("options:")
	flag.PrintDefaults()
}

const (
	sqlite = "sqlite"
	csv    = "csv"
)

func validateOutputFormats(output string, format string) error {
	switch format {
	case csv:
		return nil
	case sqlite:
		// stdout
		if output == "" || output == "-" {
			return fmt.Errorf("cannot use sqlite format and stdout together")
		}
		return nil
	default:
		return fmt.Errorf("invalid format: %s", format)
	}
}

func main() {
	flag.Usage = usage
	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}
	switch os.Args[1] {
	case "load":
		load()
	case "serve":
		serve()
	default:
		fmt.Fprintf(os.Stderr, "invalid command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
