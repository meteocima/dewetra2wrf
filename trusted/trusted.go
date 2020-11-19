package trusted

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/meteocima/dewetra2wrf/conversion"
	"github.com/meteocima/dewetra2wrf/sensor"

	"github.com/meteocima/dewetra2wrf/obsreader"
)

// InputFormat ...
type InputFormat int

// InputFormat values ...
const (
	DewetraFormat InputFormat = iota
	WundergroundFormat
	WunderHistFormat
)

func (f InputFormat) String() string {
	if f == DewetraFormat {
		return "DewetraFormat"
	}

	if f == WundergroundFormat {
		return "WundergroundFormat"
	}

	if f == DewetraFormat {
		return "WunderHistFormat"
	}

	return fmt.Sprintf("%d", int(f))
}

var headerFormat = "TOTAL = %6d, MISS. =-888888.,\n" +
	"SYNOP = %6d, METAR =      0, SHIP  =      0, BUOY  =      0, BOGUS =      0, TEMP  =      0,\n" +
	"AMDAR =      0, AIREP =      0, TAMDAR=      0, PILOT =      0, SATEM =      0, SATOB =      0,\n" +
	"GPSPW =      0, GPSZD =      0, GPSRF =      0, GPSEP =      0, SSMT1 =      0, SSMT2 =      0,\n" +
	"TOVS  =      0, QSCAT =      0, PROFL =      0, AIRSR =      0, OTHER =      0,\n" +
	"PHIC  =  40.00, XLONC = -95.00, TRUE1 =  30.00, TRUE2 =  60.00, XIM11 =   1.00, XJM11 =   1.00,\n" +
	"base_temp= 290.00, base_lapse=  50.00, PTOP  =  5000., base_pres=100000., base_tropo_pres= 20000., base_strat_temp=   215.,\n" +
	"IXC   =     60, JXC   =     90, IPROJ =      1, IDD   =      1, MAXNES=      1,\n" +
	"NESTIX=     60,\n" +
	"NESTJX=     90,\n" +
	"NUMC  =      1,\n" +
	"DIS   =  60.00,\n" +
	"NESTI =      1,\n" +
	"NESTJ =      1,\n" +
	"INFO  = PLATFORM, DATE, NAME, LEVELS, LATITUDE, LONGITUDE, ELEVATION, ID.\n" +
	"SRFC  = SLP, PW (DATA,QC,ERROR).\n" +
	"EACH  = PRES, SPEED, DIR, HEIGHT, TEMP, DEW PT, HUMID (DATA,QC,ERROR)*LEVELS.\n" +
	"INFO_FMT = (A12,1X,A19,1X,A40,1X,I6,3(F12.3,11X),6X,A40)\n" +
	"SRFC_FMT = (F12.3,I4,F7.2,F12.3,I4,F7.3)\n" +
	"EACH_FMT = (3(F12.3,I4,F7.2),11X,3(F12.3,I4,F7.2),11X,3(F12.3,I4,F7.2))\n" +
	"#------------------------------------------------------------------------------#\n"

func DownloadAndConvert(format InputFormat, dataPath string, domain sensor.Domain, date time.Time, filename string) error {
	/*
		AllSensors returns a chan of sensors read. the sensors variables are emitted in to the chan as soon as
		all sensor variables are read, and written in the out file at abs locations.
		An init function previously calculate the position of every sensor variable in the file.
	*/

	var sensorsObservations []sensor.Observation
	if format == DewetraFormat {
		obss, err := obsreader.AllSensors(dataPath, domain, date)
		if err != nil {
			return err
		}
		sensorsObservations = obss
	} else if format == WundergroundFormat {
		obss, err := obsreader.AllSensorsWund(dataPath, domain, date)
		if err != nil {
			return err
		}
		sensorsObservations = obss
	} else if format == WunderHistFormat {
		obss, err := obsreader.AllSensorsWundHistory(dataPath, domain, date)
		if err != nil {
			return err
		}
		sensorsObservations = obss

	} else {
		panic("Unknown format " + format.String())
	}

	results := make([]string, len(sensorsObservations))
	for i, result := range sensorsObservations {
		results[i] = conversion.ToWRFDA(result)
	}

	resultsS := strings.Join(results, "\n")

	header := fmt.Sprintf(headerFormat, len(results), len(results))

	return ioutil.WriteFile(filename, []byte(header+resultsS), os.FileMode(0644))

}

// Get ...
func Get(format InputFormat, data string, outputFile string, domain string, date time.Time) error {

	coords := strings.Split(domain, ",")

	MinLat, err := strconv.ParseFloat(coords[0], 64)
	if err != nil {
		return err
	}

	MaxLat, err := strconv.ParseFloat(coords[1], 64)
	if err != nil {
		return err
	}

	MinLon, err := strconv.ParseFloat(coords[2], 64)
	if err != nil {
		return err
	}

	MaxLon, err := strconv.ParseFloat(coords[3], 64)
	if err != nil {
		return err
	}
	/*
		fmt.Printf("MinLat: %f\nMinLon: %f\nMaxLat: %f\nMaxLon: %f\n",
			MinLat,
			MinLon,
			MaxLat,
			MaxLon,
		)
	*/
	return DownloadAndConvert(
		format,
		data,
		//
		// leftlon, rightlon, toplat, bottomlat
		sensor.Domain{
			MinLat: MinLat,
			MinLon: MinLon,
			MaxLat: MaxLat,
			MaxLon: MaxLon,
		},
		date,

		outputFile,
	)

}
