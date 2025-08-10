package main

import (
	"context"
	"elevation/internal/db"
	"elevation/internal/hgt"
	"elevation/internal/http"
	"elevation/internal/service"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
	if loadCmd.NArg() > 0 {
		fpath := loadCmd.Arg(0)
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
		recs := hgt.HGTRecords(records)
		err := recs.CSV(out, true)
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

func serve() {
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	serveCmd.Usage = func() {
		fmt.Printf("usage: %s serve [options] [DB FILE]\n", os.Args[0])
		fmt.Println("")
		fmt.Println("options:")
		serveCmd.PrintDefaults()
	}
	var port int
	serveCmd.IntVar(&port, "p", 8000, "port to run server on")
	var address string
	serveCmd.StringVar(&address, "a", "0.0.0.0", "interface to bind to")

	var verbose bool
	serveCmd.BoolVar(&verbose, "v", false, "enable verbose")

	err := serveCmd.Parse(os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if verbose {
		fmt.Printf("serving: %s:%d\n", address, port)
	}

	if serveCmd.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "must specifiy database file to use")
		os.Exit(1)
	}

	fpath := serveCmd.Arg(0)
	f, err := os.Open(serveCmd.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stdout, "error: could not open file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	db, err := db.NewElevationDB(fpath, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	s := service.NewElevationService(db)
	err = http.Serve(address, port, s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
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
