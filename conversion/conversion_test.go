package conversion

import (
	"testing"

	"github.com/meteocima/wund-to-ascii/download"

	"github.com/meteocima/wund-to-ascii/testutil"
	"github.com/stretchr/testify/assert"
)

func TestConvertToAscii(t *testing.T) {
	results, err := testutil.AllSensorsFromFixture(t, download.MatchDownloadedData)
	assert.NoError(t, err)
	s := ToWRFDA(results[0])
	expected := "FM-12 SYNOP  2020-03-30_18:00:00 Foggia Istituto Agrario                       1      41.469                 15.483                  0.000                 210329130_2                             \n" +
		"       0.000   0   0.10       0.000   0  0.100\n" +
		" -888888.000   0   0.10       0.600   0   0.10     292.000   0   0.10                  0.000   0   0.10      13.000   0   0.10 -888888.000   0   0.10                 75.000   0   0.10       0.000   0   0.10       0.000   0   0.10"

	assert.Equal(t, expected, s)
}
