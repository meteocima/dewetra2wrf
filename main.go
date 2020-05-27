package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/meteocima/wund-to-ascii/sensor"
	"github.com/meteocima/wund-to-ascii/trusted"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: wund-to-ascii <yyyymmddhh>")
		os.Exit(1)
	}

	date, err := time.Parse("2006010215", os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	data := "/var/local/wund-to-ascii"
	err = os.MkdirAll(data, os.FileMode(0755))
	if err != nil && !os.IsExist(err) {
		log.Fatal(err)
	}

	err = trusted.DownloadAndConvert(
		data,
		//
		// leftlon, rightlon, toplat, bottomlat
		// -19.0, 48.0, 64.0, 24.0
		sensor.Domain{MinLat: 24, MinLon: -19, MaxLat: 64, MaxLon: 48},
		date,

		path.Join(data, "ob.ascii"),
	)
	if err != nil {
		log.Fatal(err)
	}
}
