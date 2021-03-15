package testutil

/*
import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/meteocima/dewetra2wrf/sensor"
	"github.com/stretchr/testify/assert"
)

// FixtureDir return directory of fixtures
func FixtureDir(filePath string) string {
	_, currentFilePath, _, _ := runtime.Caller(0)
	//fmt.Println("currentFilePath", currentFilePath)
	result, err := filepath.Abs(filepath.Join(currentFilePath, "../../fixtures", filePath))
	if err != nil {
		panic(err)
	}
	return result
}

func MustParse(dateS string) time.Time {
	dt, err := time.Parse("200601021504", dateS)
	if err != nil {
		panic(err)
	}
	return dt
}

func MustParseISO(dateS string) time.Time {
	dt, err := time.Parse("02/01/2006 15", dateS)
	if err != nil {
		panic(err)
	}
	return dt
}

func GetResultsFile(t *testing.T, name string) []sensor.Result {
	fixturePath := FixtureDir(name)
	buff, err := ioutil.ReadFile(fixturePath)
	assert.NoError(t, err)
	var expected []sensor.Result
	err = json.Unmarshal(buff, &expected)
	assert.NoError(t, err)
	return expected
}

func SaveResultsFile(t *testing.T, name string, results interface{}) {
	fixturePath := FixtureDir(name)
	buff, err := json.Marshal(results)

	err = ioutil.WriteFile(fixturePath, buff, 0644)
	assert.NoError(t, err)
}

type Matcher func(dataPath string, domain sensor.Domain, pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater []sensor.Result) ([]sensor.Observation, error)

// AllSensorsFromFixture is
func AllSensorsFromFixture(t *testing.T, dataPath string, matchDownloadedData Matcher) ([]sensor.Observation, error) {
	pressure := GetResultsFile(t, "BAROMETRO.json")
	precipitableWater := GetResultsFile(t, "PLUVIOMETRO.json")
	relativeHumidity := GetResultsFile(t, "IGROMETRO.json")
	windSpeed := GetResultsFile(t, "ANEMOMETRO.json")
	windDirection := GetResultsFile(t, "DIREZIONEVENTO.json")
	temperature := GetResultsFile(t, "TERMOMETRO.json")
	return matchDownloadedData(dataPath, sensor.Domain{MinLat: 34, MinLon: 4, MaxLat: 47, MaxLon: 20}, pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater)
}
*/
