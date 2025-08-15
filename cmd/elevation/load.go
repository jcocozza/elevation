package main

import (
	"elevation"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func load() {
	loadCmd := flag.NewFlagSet("load", flag.ExitOnError)
	loadCmd.Usage = func() {
		fmt.Printf("usage: %s load [options] [FILE]\n", os.Args[0])
		fmt.Println("")
		fmt.Println("options:")
		loadCmd.PrintDefaults()

	}
	var tileName string
	loadCmd.StringVar(&tileName, "t", "", "tile name (e.g. 'N00E006') (default: use passed file basename. required if reading from stdin)")
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

	var in io.Reader
	useStdin := loadCmd.NArg() == 0 || (loadCmd.NArg() == 1 && loadCmd.Arg(0) == "-")
	oneArg := loadCmd.NArg() == 1 && !(loadCmd.Arg(0) == "-")

	if useStdin {
		if tileName == "" {
			fmt.Fprintln(os.Stderr, "error: must include tileName when passing to stdin")
			os.Exit(1)
		}
		in = os.Stdin
	} else if oneArg {
		fpath := loadCmd.Arg(0)
		file, err := os.Open(fpath)
		if err != nil {
			fmt.Fprintf(os.Stdout, "error: could not open file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		in = file

		if tileName == "" {
			tileName = strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))
		}
	}

	lat, lng, err := elevation.ParseTileName(tileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	records, err := elevation.ProcessHGTFile(in, lat, lng)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	writeOut(output, format, records)
}
