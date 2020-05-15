package trusted

import (
	"strings"
	"time"

	"github.com/meteocima/wund-to-ascii/conversion"
	"github.com/meteocima/wund-to-ascii/sensor"

	"github.com/meteocima/wund-to-ascii/download"
)

func DownloadAndConvert(dataPath string, from, to time.Time) (string, error) {
	download.DownloadAllSensorsTables(dataPath, sensor.DPCTrusted)

	sensorsObservations, err := download.AllSensors(dataPath, from, to)
	if err != nil {
		return "", err
	}

	results := make([]string, len(sensorsObservations))
	for i, result := range sensorsObservations {
		results[i] = conversion.ToWRFDA(result)
	}

	return strings.Join(results, "\n"), nil
}
