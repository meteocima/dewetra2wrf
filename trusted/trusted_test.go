package trusted

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/meteocima/wund-to-ascii/sensor"

	"github.com/meteocima/wund-to-ascii/testutil"
	"github.com/stretchr/testify/assert"
)

var header = "MISS. =-888888.,\n" +
	"SYNOP =  10000, METAR =   2416, SHIP  =     54, BUOY  =    341, BOGUS =      0, TEMP  =     86,\n" +
	"AMDAR =     14, AIREP =    205, TAMDAR=      0, PILOT =     85, SATEM =    106, SATOB =   2556,\n" +
	"GPSPW =    187, GPSZD =      0, GPSRF =      3, GPSEP =      0, SSMT1 =      0, SSMT2 =      0,\n" +
	"TOVS  =      0, QSCAT =   2190, PROFL =     61, AIRSR =      0, OTHER =      0,\n" +
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

func TestDownloadPrecipitableWater(t *testing.T) {
	data := testutil.FixtureDir("testanagr")
	os.MkdirAll(data, os.FileMode(0755))
	defer os.RemoveAll(data)
	results, err := DownloadAndConvert(
		data,
		// LIGURIA sensor.Domain{MinLat: 43, MinLon: 7, MaxLat: 44, MaxLon: 10},
		sensor.Domain{MinLat: 34, MinLon: 4, MaxLat: 47, MaxLon: 20},
		time.Date(2020, 5, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 5, 10, 1, 0, 0, 0, time.UTC),
	)

	resultsS := strings.Join(results, "\n")

	totalS := fmt.Sprintf("TOTAL = %6d, ", len(results))

	assert.NoError(t, err)
	err = ioutil.WriteFile("/home/parroit/dpc.txt", []byte(totalS+header+resultsS), os.FileMode(0644))
	assert.NoError(t, err)
	assert.Equal(t, "", results)

}
