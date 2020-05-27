package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

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

	err = trusted.Get(data, path.Join(data, "ob.ascii"), date)

	if err != nil {
		log.Fatal(err)
	}
}
