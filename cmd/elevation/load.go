package main

import (
	"context"
	"elevation"
	"elevation/pkg/db"
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

	records, err := elevation.ProcessHGT(in, lat, lng)
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
		err := elevation.HGTToCSV(out, true, records)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	case "sqlite":
		db, err := db.NewElevationDB(output, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if err = db.CreateRecords(context.TODO(), records); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "invalid format: %s\n", format)
		os.Exit(1)
	}
}
