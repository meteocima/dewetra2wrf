package obsreader

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/meteocima/dewetra2wrf/elevations"
	"github.com/meteocima/dewetra2wrf/types"
)

// WundCurrentObsReader reads
// observations from JSON files as returned
// from wunderground 'current' web API.
type WundCurrentObsReader struct{}

// ReadAll implements ObsReader for WundCurrentObsReader
func (r WundCurrentObsReader) ReadAll(dataPath string, domain types.Domain, date time.Time) ([]types.Observation, error) {
	dateDir := filepath.Join(dataPath, date.Format("2006010215"))
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
		var obs types.Observation
		err = json.Unmarshal(obsBuf, &obs)
		if err != nil {
			return nil, err
		}
		if obs.Lat <= domain.MaxLat && obs.Lat >= domain.MinLat &&
			obs.Lon <= domain.MaxLon && obs.Lon >= domain.MinLon {

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
	return observations, nil
}
