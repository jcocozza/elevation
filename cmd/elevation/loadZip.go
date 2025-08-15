package main

import (
	"archive/zip"
	"elevation"
	"flag"
	"fmt"
	"os"
)

func loadZip() {
	loadCmd := flag.NewFlagSet("load-zip", flag.ExitOnError)
	loadCmd.Usage = func() {
		fmt.Printf("usage: %s load-zip [options] [FILE]\n", os.Args[0])
		fmt.Println("")
		fmt.Println("does not read from stdin")
		fmt.Println("")
		fmt.Println("options:")
		loadCmd.PrintDefaults()
	}

	var output string
	loadCmd.StringVar(&output, "o", "", "file name to output (default: stdout)")
	var format string
	loadCmd.StringVar(&format, "f", "csv", "output format (options: sqlite, csv)")

	err := loadCmd.Parse(os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if err := validateOutputFormats(output, format); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if loadCmd.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "missing path to zip file")
		os.Exit(1)
	}

	fpath := loadCmd.Arg(0)
	in, err := zip.OpenReader(fpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	records, err := elevation.ProcessZippedHGT(in)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	writeOut(output, format, records)
}
