package wunderground

import (
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/meteocima/dewetra2wrf/sensor"
	"github.com/stretchr/testify/assert"
)

// DataDir return directory of fixtures
func DataDir(filePath string) string {
	_, currentFilePath, _, _ := runtime.Caller(0)
	result, err := filepath.Abs(filepath.Join(currentFilePath, "../../../data", filePath))
	if err != nil {
		panic(err)
	}
	return result
}

func TestReadWundergroundFormat(t *testing.T) {
	obss, err := Read(DataDir("wundformat"), sensor.Domain{}, time.Date(2020, 7, 15, 6, 0, 0, 0, time.UTC))
	assert.NoError(t, err)
	assert.NotNil(t, obss)

}
