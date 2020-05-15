package trusted

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/meteocima/wund-to-ascii/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDownloadPrecipitableWater(t *testing.T) {
	data := testutil.FixtureDir("testanagr")
	os.MkdirAll(data, os.FileMode(0755))
	defer os.RemoveAll(data)
	results, err := DownloadAndConvert(data, time.Date(2020, 5, 10, 0, 0, 0, 0, time.UTC), time.Date(2020, 5, 10, 1, 0, 0, 0, time.UTC))
	assert.NoError(t, err)
	err = ioutil.WriteFile("/home/parroit/dpc.txt", []byte(results), os.FileMode(0644))
	assert.NoError(t, err)
	assert.Equal(t, "", results)

}
