package obsreader

/*
import (
	"encoding/json"
	"io/ioutil"
	"math"
	"testing"

	"github.com/meteocima/dewetra2wrf/types"
	"github.com/meteocima/dewetra2wrf/testutil"
	"github.com/stretchr/testify/assert"
)


func TestMatchDownloadedData(t *testing.T) {
	pressure := testutil.GetResultsFile(t, "BAROMETRO.json")
	precipitableWater := testutil.GetResultsFile(t, "PLUVIOMETRO.json")
	relativeHumidity := testutil.GetResultsFile(t, "IGROMETRO.json")
	windSpeed := testutil.GetResultsFile(t, "ANEMOMETRO.json")
	windDirection := testutil.GetResultsFile(t, "DIREZIONEVENTO.json")
	temperature := testutil.GetResultsFile(t, "TERMOMETRO.json")
	results, err := mergeObservations(testutil.FixtureDir("anagr"), types.Domain{
		MinLat: -180,
		MaxLat: 180,
		MinLon: -90,
		MaxLon: 90,
	}, pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater)
	assert.NoError(t, err)

	assert.Equal(t, 375, len(results))
	assert.Equal(t, results[0].StationID, "210329130_2")
	assert.Equal(t, results[0].StationName, "Foggia Istituto Agrario")
	assert.Equal(t, results[0].ObsTimeUtc, testutil.MustParseISO("2020-03-30T18:00:00Z"))
	assert.Equal(t, results[0].Lat, 41.469000)
	assert.Equal(t, results[0].Lon, 15.483167)
	assert.Equal(t, results[0].HumidityAvg, types.Value(75.00000))
	assert.Equal(t, results[0].WinddirAvg, types.Value(292.00000))
	assert.Equal(t, results[0].Metric.TempAvg, types.Value(13.00000))
	assert.Equal(t, results[0].Metric.WindspeedAvg, types.Value(0.60000))
	assert.True(t, math.IsNaN(float64(results[0].Metric.Pressure)))
	assert.Equal(t, results[0].Metric.PrecipTotal, types.Value(0.00000))

}

func TestSensorValueUnmarshalNaN(t *testing.T) {
	buff, err := ioutil.ReadFile(testutil.FixtureDir("expected-download-results.json"))
	assert.NoError(t, err)

	var observations []types.Observation
	err = json.Unmarshal(buff, &observations)
	assert.NoError(t, err)

	assert.True(t, math.IsNaN(observations[0].Metric.Pressure.AsFloat()))

}
*/
