package main

import (
	"elevation/internal/hgt"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func usage() {
	fmt.Printf("usage: %s [options] [FILE]\n", os.Args[0])
	fmt.Println("")
	fmt.Println("read from stdin and display the results in a selectable list")
	fmt.Println("")
	fmt.Println("options:")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage

	var tileName string
	flag.StringVar(&tileName, "t", "", "tile name (e.g. 'N00E006') (default: use passed file basename. required if reading from stdin)")

	var output string
	flag.StringVar(&output, "o", "", "file name to output (default: stdout)")

	var format string
	flag.StringVar(&format, "f", "csv", "output format (options: sqlite, csv)")

	flag.Parse()

	var in io.Reader
	if flag.NArg() > 0 {
		fpath := flag.Arg(0)
		file, err := os.Open(fpath)
		if err != nil {
			fmt.Fprintf(os.Stdout, "error: could not open file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		in = file

		if tileName == "" {
			tileName = strings.TrimSuffix(fpath, filepath.Ext(fpath))
		}

	} else {
		if tileName == "" {
			fmt.Fprintln(os.Stderr, "error: must include tileName when passing to stdin")
			os.Exit(1)
		}
		in = os.Stdin
	}

	lat, lng, err := hgt.ParseTileName(tileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	records, err := hgt.ProcessHGT(in, lat, lng)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var out io.Writer
	if output == "" || output == "-" {
		out = os.Stdout
	} else {
		f, err := os.Create(output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		out = f
	}

	switch format {
	case "csv":
		err := hgt.HGTRecords(records).CSV(out, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "sqlite": // TODO: sqlite output
	default:
		fmt.Fprintf(os.Stderr, "invalid format: %s\n", format)
		os.Exit(1)
	}
}
