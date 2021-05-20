package obsreader

import (
	"encoding/json"
	"io/ioutil"
	"math"
	"path/filepath"
	"time"

	"github.com/meteocima/dewetra2wrf/elevations"
	"github.com/meteocima/dewetra2wrf/types"
)

// WundHistObsReader reads
// observations from JSON files as returned
// from wunderground 'historical' web API.
type WundHistObsReader struct{}

// ReadAll implements ObsReader for WundHistObsReader
func (r WundHistObsReader) ReadAll(dataPath string, domain types.Domain, date time.Time) ([]types.Observation, error) {
	var dateDir string
	if !date.IsZero() {
		dateDir = filepath.Join(dataPath, date.Format("20060102"))
	} else {
		dateDir = dataPath
	}

	files, err := ioutil.ReadDir(dateDir)
	if err != nil {
		return nil, err
	}
	observations := []types.Observation{}

	for _, f := range files {
		obsBuf, err := ioutil.ReadFile(filepath.Join(dateDir, f.Name()))
		if err != nil {
			return nil, err
		}
		var obsList struct {
			Observations []types.Observation
		}

		err = json.Unmarshal(obsBuf, &obsList)
		if err != nil {
			return nil, err
		}

		if len(obsList.Observations) == 0 {
			continue
		}
		var obs types.Observation = obsList.Observations[0]

		if date.IsZero() {
			for _, obs := range obsList.Observations {
				obs.Elevation = elevations.GetFromCoord(obs.Lat, obs.Lon)
				obs.StationName = obs.StationID
				obs.Metric.Pressure = types.Value((obs.Metric.PressureMax + obs.Metric.PressureMin) / 2)
				// convert temperatures from °celsius to °kelvin
				obs.Metric.TempAvg += 273.15
				// convert wind speed from km/h into m/s
				obs.Metric.WindspeedAvg *= 0.277778
				// convert pressure from mbar into Pa
				obs.Metric.Pressure *= 100

				observations = append(observations, obs)
			}
		} else {
			if obs.Lat <= domain.MaxLat && obs.Lat >= domain.MinLat &&
				obs.Lon <= domain.MaxLon && obs.Lon >= domain.MinLon {

				minDeltaMin := 30.0
				for _, o := range obsList.Observations {

					delta := math.Abs(date.Sub(o.ObsTimeUtc).Minutes())
					if delta < minDeltaMin {
						minDeltaMin = delta
						obs = o
					}
					if delta == 0.0 {
						break
					}
				}

				obs.Elevation = elevations.GetFromCoord(obs.Lat, obs.Lon)
				obs.StationName = obs.StationID
				obs.Metric.Pressure = types.Value((obs.Metric.PressureMax + obs.Metric.PressureMin) / 2)
				// convert temperatures from °celsius to °kelvin
				obs.Metric.TempAvg += 273.15
				// convert wind speed from km/h into m/s
				obs.Metric.WindspeedAvg *= 0.277778
				// convert pressure from mbar into Pa
				obs.Metric.Pressure *= 100

				observations = append(observations, obs)
			}
		}

	}
	return observations, nil
}
