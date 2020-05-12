package dewetra

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToAscii(t *testing.T) {
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
