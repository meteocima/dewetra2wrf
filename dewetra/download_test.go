package dewetra

import (
	"encoding/json"
	"io/ioutil"
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

func mustParse(dateS string) time.Time {
	dt, err := time.Parse("200601021504", dateS)
	if err != nil {
		panic(err)
	}
	return dt
}

func getResultsFile(t *testing.T, name string) SensorsResult {
	fixturePath := testutil.FixtureDir(name)
	buff, err := ioutil.ReadFile(fixturePath)
	assert.NoError(t, err)
	var expected SensorsResult
	err = json.Unmarshal(buff, &expected)
	assert.NoError(t, err)
	return expected
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
	resultsBuff, err := json.MarshalIndent(results, " ", " ")
	assert.NoError(t, err)
	//fmt.Println(string(resultsBuff))
	_ = resultsBuff
}
