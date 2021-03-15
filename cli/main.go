package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/meteocima/dewetra2wrf"
)

func main() {
	format := flag.String("format", ".", "format of input files (DEWETRA or WUNDERGROUND)")
	input := flag.String("input", ".", "where to read input files")
	outfile := flag.String("outfile", "./out", "where to save converted file")
	domainS := flag.String("domain", "", "domain to filter stations to download [MinLat,MaxLat,MinLon,MaxLon]")
	dateS := flag.String("date", "", "date and hour of the data to download [YYYYMMDDHH]")

	flag.Parse()

	if *domainS == "" || *dateS == "" {
		flag.Usage()
		os.Exit(1)
	}

	date, err := time.Parse("2006010215", *dateS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		flag.Usage()
		os.Exit(1)
	}

	var form dewetra2wrf.InputFormat
	form.FromString(*format)

	err = dewetra2wrf.Convert(form, *input, *domainS, date, *outfile)

	if err != nil {
		log.Fatal(err)
	}
}
