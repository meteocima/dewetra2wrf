package conversion

import (
	"strings"
	"testing"
	"time"

	"github.com/meteocima/dewetra2wrf/types"
	"github.com/stretchr/testify/assert"
)

var testobs = types.Observation{
	Elevation:   1234,
	StationID:   "210329130_2",
	StationName: "Foggia Istituto Agrario",
	ObsTimeUtc:  time.Date(2020, 3, 30, 18, 1, 2, 0, time.UTC),
	Lat:         41.469,
	Lon:         15.483,
	HumidityAvg: 5,
	WinddirAvg:  6,
	Metric: types.ObservationMetric{
		TempAvg:      7,
		WindspeedAvg: 8,
		Pressure:     9,
		PrecipTotal:  10,
	},
}

func TestConvertToAscii(t *testing.T) {

	actual := ToWRFASCII(testobs)

	expected := []string{
		"FM-12 SYNOP  2020-03-30_18:01:02 FoggiaXIstitutoXAgrario                       1      41.469                 15.483               1234.000                 XXXXXXXXXXX                             ",
		" -888888.000 -88  99.99 -888888.000 -88 99.990",
		"       9.000   0   1.00       8.000   0   1.00       6.000   0   3.00            -888888.000 -88 999.99       7.000   0   1.00 -888888.000 -88   1.00                  5.000   0   2.00",
	}
	for i, l := range strings.Split(actual, "\n") {
		assert.Equal(t, expected[i], l, "lines %d differs", i)

	}
}
