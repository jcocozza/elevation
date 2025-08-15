package main

import (
	"archive/zip"
	"elevation"
	"elevation/internal/util"
	"flag"
	"fmt"
	"os"
)

func queryZip() {
	queryCmd := flag.NewFlagSet("query-zip", flag.ExitOnError)
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

	if err := util.ValidateLatLng(lat, lng); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	fpath := queryCmd.Arg(0)
	in, err := zip.OpenReader(fpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	elevation, err := elevation.GetElevationFromZip(in, lat, lng, elevation.SRTM1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(elevation)
}
