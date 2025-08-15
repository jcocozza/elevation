package main

import (
	"elevation"
	"elevation/pkg/db"
	"fmt"
	"io"
	"os"
)

func writeOut(output string, format string, records []elevation.HGTRecord) {
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
		d, err := db.NewElevationDB(output, false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if err = db.CreateRecords(d.(*db.ElevationSQLiteDB), records); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if err = db.CreateFinalTable(d.(*db.ElevationSQLiteDB)); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "invalid format: %s\n", format)
		os.Exit(1)
	}

}
