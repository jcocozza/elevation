package main

import (
	"elevation/pkg/db"
	"elevation/pkg/server"
	"elevation/pkg/service"
	"flag"
	"fmt"
	"os"
)

func serve() {
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	serveCmd.Usage = func() {
		fmt.Printf("usage: %s serve [options] [DIR]\n", os.Args[0])
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
	err = server.Serve(address, port, s)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}
