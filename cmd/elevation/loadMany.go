package main

import (
	"elevation"
	"elevation/pkg/db"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func loadMany() {
	loadCmd := flag.NewFlagSet("load-many", flag.ExitOnError)
	loadCmd.Usage = func() {
		fmt.Printf("usage: %s load-many [options] [FILES...]\n", os.Args[0])
		fmt.Println("")
		fmt.Println("assumes files are names so that tile name can be extracted from it")
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

	args := loadCmd.Args()

	files := []string{}
	for _, arg := range args {
		matches, err := filepath.Glob(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid pattern: %v\n", err)
			os.Exit(1)
		}
		// no matches, this is an actual file
		if matches == nil {
			files = append(files, arg)
			continue
		}
		files = append(files, matches...)
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

	var d db.ElevationDB
	if format == sqlite {
		d, err = db.NewElevationDB(output, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	// process all the files
	for i, fpath := range files {
		tileName := strings.TrimSuffix(filepath.Base(fpath), filepath.Ext(fpath))
		lat, lng, err := elevation.ParseTileName(tileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		f, err := os.Open(fpath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		records, err := elevation.ProcessHGTFile(f, lat, lng)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		switch format {
		case "csv":
			includeHeader := i == 0
			err := elevation.HGTToCSV(out, includeHeader, records)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
		case "sqlite":
			if err = db.CreateRecords(d.(*db.ElevationSQLiteDB), records); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}

		default:
			fmt.Fprintf(os.Stderr, "invalid format: %s\n", format)
			os.Exit(1)
		}
	}

	if format == sqlite {
		fmt.Println("records created")
		if err = db.CreateFinalTable(d.(*db.ElevationSQLiteDB)); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}
}
