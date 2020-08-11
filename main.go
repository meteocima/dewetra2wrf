package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/meteocima/dewetra2wrf/trusted"
)

func main() {
	input := flag.String("input", ".", "where to read input files")
	outfile := flag.String("outfile", "./out", "where to save converted file")
	domainS := flag.String("domain", "", "domain to filter stations to download [MinLat,MaxLat,MinLon,MaxLon]")
	dateS := flag.String("date", "", "date and hour of the data to download [YYYYMMDDHH]")

	flag.Parse()

	if *domainS == "" || *dateS == "" {
		flag.Usage()
		os.Exit(1)
	}

	parts := strings.Split(*domainS, ",")
	numParts := make([]float64, 4)
	for i := 0; i < 4; i++ {
		n, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			flag.Usage()
			os.Exit(1)
		}
		numParts[i] = n
	}

	date, err := time.Parse("2006010215", *dateS)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		flag.Usage()
		os.Exit(1)
	}

	err = trusted.Get(*input, *outfile, date)

	if err != nil {
		log.Fatal(err)
	}
}
