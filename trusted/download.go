package trusted

import (
	"time"

	"github.com/meteocima/wund-to-ascii/download"
)

func Download(from, to time.Time) (string, error) {
	var ids []string
	results, err := download.AllSensors("anagr", ids, from, to)
	if err != nil {
		return "", err
	}
	_ = results
	return "string(results)", nil
}
