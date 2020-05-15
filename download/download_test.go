package download

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"testing"
	"time"

	"github.com/meteocima/wund-to-ascii/sensor"
	"github.com/meteocima/wund-to-ascii/testutil"
	"github.com/stretchr/testify/assert"
)

/*
{
	"sensorClass": "PLUVIOMETRO",
    "from": "202003301800",
    "to": "202003310001",
    "ids": ["479124536_2", "479072112_2"]
}
*/

func TestDownloadPrecipitableWater(t *testing.T) {
	result, err := downloadPrecipitableWater(
		[]string{"-2147444447_2", "268445238_2"},
		time.Date(2020, 3, 30, 18, 0, 0, 0, time.UTC),
		time.Date(2020, 3, 31, 0, 1, 0, 0, time.UTC),
	)
	assert.NoError(t, err)
	assert.Equal(t, testutil.GetResultsFile(t, "PLUVIOMETRO.json"), result)
}

func TestDownloadRelativeHumidity(t *testing.T) {
	result, err := downloadRelativeHumidity(
		[]string{"210329130_2", "9784_2"},
		time.Date(2020, 3, 30, 18, 0, 0, 0, time.UTC),
		time.Date(2020, 3, 31, 0, 1, 0, 0, time.UTC),
	)
	assert.NoError(t, err)
	assert.Equal(t, testutil.GetResultsFile(t, "IGROMETRO.json"), result)
}

func TestDownloadWindSpeed(t *testing.T) {
	result, err := downloadWindSpeed(
		[]string{"210329129_2", "39011_2"},
		time.Date(2020, 3, 30, 18, 0, 0, 0, time.UTC),
		time.Date(2020, 3, 31, 0, 1, 0, 0, time.UTC),
	)
	assert.NoError(t, err)
	assert.Equal(t, testutil.GetResultsFile(t, "ANEMOMETRO.json"), result)

}

func TestDownloadWindDirection(t *testing.T) {
	result, err := downloadWindDirection(
		[]string{"210329131_2", "39010_2"},
		time.Date(2020, 3, 30, 18, 0, 0, 0, time.UTC),
		time.Date(2020, 3, 31, 0, 1, 0, 0, time.UTC),
	)
	assert.NoError(t, err)
	assert.Equal(t, testutil.GetResultsFile(t, "DIREZIONEVENTO.json"), result)
}

func TestDownloadTemperature(t *testing.T) {
	result, err := downloadTemperature(
		[]string{"39202_2", "9781_2"},
		time.Date(2020, 3, 30, 18, 0, 0, 0, time.UTC),
		time.Date(2020, 3, 31, 0, 1, 0, 0, time.UTC),
	)
	assert.NoError(t, err)
	assert.Equal(t, testutil.GetResultsFile(t, "TERMOMETRO.json"), result)
}

func TestDownloadPressure(t *testing.T) {
	result, err := downloadPressure(
		[]string{"9783_2", "7521_2"},
		time.Date(2020, 3, 30, 18, 0, 0, 0, time.UTC),
		time.Date(2020, 3, 31, 0, 1, 0, 0, time.UTC),
	)
	assert.NoError(t, err)
	assert.Equal(t, testutil.GetResultsFile(t, "BAROMETRO.json"), result)

}

func TestMatchDownloadedData(t *testing.T) {
	pressure := testutil.GetResultsFile(t, "BAROMETRO.json")
	precipitableWater := testutil.GetResultsFile(t, "PLUVIOMETRO.json")
	relativeHumidity := testutil.GetResultsFile(t, "IGROMETRO.json")
	windSpeed := testutil.GetResultsFile(t, "ANEMOMETRO.json")
	windDirection := testutil.GetResultsFile(t, "DIREZIONEVENTO.json")
	temperature := testutil.GetResultsFile(t, "TERMOMETRO.json")
	results, err := MatchDownloadedData(pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater)
	assert.NoError(t, err)

	assert.Equal(t, 375, len(results))
	assert.Equal(t, results[0].StationID, "210329130_2")
	assert.Equal(t, results[0].StationName, "Foggia Istituto Agrario")
	assert.Equal(t, results[0].ObsTimeUtc, testutil.MustParseISO("2020-03-30T18:00:00Z"))
	assert.Equal(t, results[0].Lat, 41.469000)
	assert.Equal(t, results[0].Lon, 15.483167)
	assert.Equal(t, results[0].HumidityAvg, sensor.Value(75.00000))
	assert.Equal(t, results[0].WinddirAvg, sensor.Value(292.00000))
	assert.Equal(t, results[0].Metric.TempAvg, sensor.Value(13.00000))
	assert.True(t, math.IsNaN(float64(results[0].Metric.DewptAvg)))
	assert.Equal(t, results[0].Metric.WindspeedAvg, sensor.Value(0.60000))
	assert.True(t, math.IsNaN(float64(results[0].Metric.Pressure)))
	assert.Equal(t, results[0].Metric.PrecipTotal, sensor.Value(0.00000))

}

func TestSensorValueUnmarshalNaN(t *testing.T) {
	buff, err := ioutil.ReadFile(testutil.FixtureDir("expected-download-results.json"))
	assert.NoError(t, err)

	var observations []sensor.Observation
	err = json.Unmarshal(buff, &observations)
	assert.NoError(t, err)

	assert.True(t, math.IsNaN(observations[0].Metric.DewptAvg.AsFloat()))
	//fmt.Println(string(resultsBuff))

}
