package main

import (
	"context"
	"elevation"
	"elevation/pkg/db"
	"elevation/pkg/service"
	"flag"
	"fmt"
	"os"
)

func queryZipMany() {
	queryCmd := flag.NewFlagSet("query-zip-many", flag.ExitOnError)
	queryCmd.Usage = func() {
		fmt.Printf("usage: %s query-zip [options] [FILE]\n", os.Args[0])
		fmt.Println("")
		fmt.Println("options:")
		queryCmd.PrintDefaults()
	}

	var lat float64
	queryCmd.Float64Var(&lat, "lat", 200, "the latitute to query")
	var lng float64
	queryCmd.Float64Var(&lng, "lng", 200, "the longitude to query")

	err := queryCmd.Parse(os.Args[2:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if queryCmd.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "expected path to file")
		os.Exit(1)
	}
	fpath := queryCmd.Arg(0)
	d, err := db.NewElevationDB(fpath, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	s := service.NewElevationService(d)
	rec, err := s.GetPointElevation(context.TODO(), lat, lng, elevation.SRTM3, service.NearestNeighbor)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(rec)
}
