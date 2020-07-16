package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/meteocima/dewetra2wrf/trusted"
)

func main() {
	//outdir := flag.String("outdir", ".", "where to save downloaded files")
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

	data := "/var/local/dewetra2wrf"

	err = trusted.Get(data, path.Join(data, "ob.ascii"), date)

	if err != nil {
		log.Fatal(err)
	}
}
