package dewetra

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"testing"
	"time"

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

func getResultsFile(t *testing.T, name string) SensorsResult {
	fixturePath := testutil.FixtureDir(name)
	buff, err := ioutil.ReadFile(fixturePath)
	assert.NoError(t, err)
	var expected SensorsResult
	err = json.Unmarshal(buff, &expected)
	assert.NoError(t, err)
	return expected
}

func saveResultsFile(t *testing.T, name string, results interface{}) {
	fixturePath := testutil.FixtureDir(name)
	buff, err := json.Marshal(results)

	err = ioutil.WriteFile(fixturePath, buff, 0644)
	assert.NoError(t, err)
}

func TestDownloadPrecipitableWater(t *testing.T) {
	result, err := downloadPrecipitableWater(
		[]string{"-2147444447_2", "268445238_2"},
		time.Date(2020, 3, 30, 18, 0, 0, 0, time.UTC),
		time.Date(2020, 3, 31, 0, 1, 0, 0, time.UTC),
	)
	assert.NoError(t, err)
	assert.Equal(t, getResultsFile(t, "PLUVIOMETRO.json"), result)
}

func TestDownloadRelativeHumidity(t *testing.T) {
	result, err := downloadRelativeHumidity(
		[]string{"210329130_2", "9784_2"},
		time.Date(2020, 3, 30, 18, 0, 0, 0, time.UTC),
		time.Date(2020, 3, 31, 0, 1, 0, 0, time.UTC),
	)
	assert.NoError(t, err)
	assert.Equal(t, getResultsFile(t, "IGROMETRO.json"), result)
}

func TestDownloadWindSpeed(t *testing.T) {
	result, err := downloadWindSpeed(
		[]string{"210329129_2", "39011_2"},
		time.Date(2020, 3, 30, 18, 0, 0, 0, time.UTC),
		time.Date(2020, 3, 31, 0, 1, 0, 0, time.UTC),
	)
	assert.NoError(t, err)
	assert.Equal(t, getResultsFile(t, "ANEMOMETRO.json"), result)

}

func TestDownloadWindDirection(t *testing.T) {
	result, err := downloadWindDirection(
		[]string{"210329131_2", "39010_2"},
		time.Date(2020, 3, 30, 18, 0, 0, 0, time.UTC),
		time.Date(2020, 3, 31, 0, 1, 0, 0, time.UTC),
	)
	assert.NoError(t, err)
	assert.Equal(t, getResultsFile(t, "DIREZIONEVENTO.json"), result)
}

func TestDownloadTemperature(t *testing.T) {
	result, err := downloadTemperature(
		[]string{"39202_2", "9781_2"},
		time.Date(2020, 3, 30, 18, 0, 0, 0, time.UTC),
		time.Date(2020, 3, 31, 0, 1, 0, 0, time.UTC),
	)
	assert.NoError(t, err)
	assert.Equal(t, getResultsFile(t, "TERMOMETRO.json"), result)
}

func TestDownloadPressure(t *testing.T) {
	result, err := downloadPressure(
		[]string{"9783_2", "7521_2"},
		time.Date(2020, 3, 30, 18, 0, 0, 0, time.UTC),
		time.Date(2020, 3, 31, 0, 1, 0, 0, time.UTC),
	)
	assert.NoError(t, err)
	assert.Equal(t, getResultsFile(t, "BAROMETRO.json"), result)

}

func TestMatchDownloadedData(t *testing.T) {
	pressure := getResultsFile(t, "BAROMETRO.json")
	precipitableWater := getResultsFile(t, "PLUVIOMETRO.json")
	relativeHumidity := getResultsFile(t, "IGROMETRO.json")
	windSpeed := getResultsFile(t, "ANEMOMETRO.json")
	windDirection := getResultsFile(t, "DIREZIONEVENTO.json")
	temperature := getResultsFile(t, "TERMOMETRO.json")
	results, err := matchDownloadedData(pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater)
	assert.NoError(t, err)

	assert.Equal(t, 375, len(results))
	assert.Equal(t, results[0].StationID, "210329130_2")
	assert.Equal(t, results[0].StationName, "Foggia Istituto Agrario")
	assert.Equal(t, results[0].ObsTimeUtc, testutil.MustParseISO("2020-03-30T18:00:00Z"))
	assert.Equal(t, results[0].Lat, 41.469000)
	assert.Equal(t, results[0].Lon, 15.483167)
	assert.Equal(t, results[0].HumidityAvg, SensorData(75.00000))
	assert.Equal(t, results[0].WinddirAvg, SensorData(292.00000))
	assert.Equal(t, results[0].Metric.TempAvg, SensorData(13.00000))
	assert.True(t, math.IsNaN(float64(results[0].Metric.DewptAvg)))
	assert.Equal(t, results[0].Metric.WindspeedAvg, SensorData(0.60000))
	assert.True(t, math.IsNaN(float64(results[0].Metric.Pressure)))
	assert.Equal(t, results[0].Metric.PrecipTotal, SensorData(0.00000))

}

func TestSensorDataUnmarshalNaN(t *testing.T) {
	buff, err := ioutil.ReadFile(testutil.FixtureDir("expected-download-results.json"))
	assert.NoError(t, err)

	var observations []Observation
	err = json.Unmarshal(buff, &observations)
	assert.NoError(t, err)

	assert.True(t, math.IsNaN(observations[0].Metric.DewptAvg.AsFloat()))
	//fmt.Println(string(resultsBuff))

}
