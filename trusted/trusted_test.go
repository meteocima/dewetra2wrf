package trusted

import (
	"os"
	"testing"
	"time"

	"github.com/meteocima/wund-to-ascii/sensor"

	"github.com/meteocima/wund-to-ascii/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDownloadPrecipitableWater(t *testing.T) {
	data := testutil.FixtureDir("testanagr")
	os.MkdirAll(data, os.FileMode(0755))
	defer os.RemoveAll(data)
	err := DownloadAndConvert(
		data,
		// LIGURIA sensor.Domain{MinLat: 43, MinLon: 7, MaxLat: 44, MaxLon: 10},
		sensor.Domain{MinLat: 34, MinLon: 4, MaxLat: 47, MaxLon: 20},
		time.Date(2020, 5, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2020, 5, 10, 1, 0, 0, 0, time.UTC),
		"/home/parroit/dpc.txt",
	)

	assert.NoError(t, err)

}
