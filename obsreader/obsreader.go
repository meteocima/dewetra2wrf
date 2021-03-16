package obsreader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"path"
	"time"

	"github.com/meteocima/dewetra2wrf/elevations"
	"github.com/meteocima/dewetra2wrf/types"
)

type ObsReader interface {
	ReadAll(dataPath string, domain types.Domain, date time.Time) ([]types.Observation, error)
}

/*

> * wind speed
ANEMOMETRO
> * wind direction
DIREZIONEVENTO
> * dewpoint temperature
Non esiste, puoi calcolarla
> * temperature
TERMOMETRO
> * relative humidity
IGROMETRO
> * precipitable water
PLUVIOMETRO

*/
/*
type sensorReqBody struct {
	SensorClass string   `json:"sensorClass"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	Ids         []string `json:"ids"`
}
*/
type sensorData struct {
	SensorID string
	Timeline []string
	Values   []float64
}

type sensorAnag struct {
	ID                  string
	Name                string
	MU                  string
	Lng, Lat, Elevation float64
}

func observationIsLess(this, that types.Result) bool {
	if that.SortKey == "zzzzzzzzzzzz" {
		return true
	}
	if this.SortKey == "zzzzzzzzzzzz" {
		return false
	}
	if this.SortKey == that.SortKey {
		return this.At.Unix() < that.At.Unix()
	}
	return this.SortKey < that.SortKey
}

func minObservation(results ...types.Result) types.Result {
	min := types.Result{SortKey: "zzzzzzzzzzzz"}
	for _, result := range results {
		if observationIsLess(result, min) {
			min = result
		}
	}
	return min
}

/*
type byName struct {
	ids   []string
	table map[string]sensorAnag
}

// Len is the number of elements in the collection.
func (bn byName) Len() int {
	return len(bn.ids)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (bn byName) Less(i, j int) bool {
	idI := bn.ids[i]
	idJ := bn.ids[j]
	return bn.table[idI].Name < bn.table[idJ].Name
}

// Swap swaps the elements with indexes i and j.
func (bn byName) Swap(i, j int) {
	save := bn.ids[i]
	bn.ids[i] = bn.ids[j]
	bn.ids[j] = save
}
*/

// MergeObservations is
func MergeObservations(dataPath string, domain types.Domain, pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater []types.Result) ([]types.Observation, error) {
	pressureIdx := 0
	relativeHumidityIdx := 0
	temperatureIdx := 0
	windDirectionIdx := 0
	windSpeedIdx := 0
	precipitableWaterIdx := 0

	results := []types.Observation{}

	sensorsTable, err := openCompleteSensorsMap(dataPath, domain)
	if err != nil {
		return nil, err
	}

	for {
		var pressureItem types.Result
		if len(pressure) > pressureIdx {
			pressureItem = pressure[pressureIdx]
		} else {
			pressureItem.SortKey = "zzzzzzzzzzzz"
		}

		var relativeHumidityItem types.Result
		if len(relativeHumidity) > relativeHumidityIdx {
			relativeHumidityItem = relativeHumidity[relativeHumidityIdx]
		} else {
			relativeHumidityItem.SortKey = "zzzzzzzzzzzz"
		}

		var temperatureItem types.Result
		if len(temperature) > temperatureIdx {
			temperatureItem = temperature[temperatureIdx]
		} else {
			temperatureItem.SortKey = "zzzzzzzzzzzz"
		}

		var windDirectionItem types.Result
		if len(windDirection) > windDirectionIdx {
			windDirectionItem = windDirection[windDirectionIdx]
		} else {
			windDirectionItem.SortKey = "zzzzzzzzzzzz"
		}

		var windSpeedItem types.Result
		if len(windSpeed) > windSpeedIdx {
			windSpeedItem = windSpeed[windSpeedIdx]
		} else {
			windSpeedItem.SortKey = "zzzzzzzzzzzz"
		}

		var precipitableWaterItem types.Result
		if len(precipitableWater) > precipitableWaterIdx {
			precipitableWaterItem = precipitableWater[precipitableWaterIdx]
		} else {
			precipitableWaterItem.SortKey = "zzzzzzzzzzzz"
		}

		if relativeHumidityItem.SortKey == "zzzzzzzzzzzz" &&
			temperatureItem.SortKey == "zzzzzzzzzzzz" &&
			windDirectionItem.SortKey == "zzzzzzzzzzzz" &&
			windSpeedItem.SortKey == "zzzzzzzzzzzz" &&
			precipitableWaterItem.SortKey == "zzzzzzzzzzzz" &&
			pressureItem.SortKey == "zzzzzzzzzzzz" {
			break
		}

		minItem := minObservation(pressureItem, relativeHumidityItem, temperatureItem, windDirectionItem, windSpeedItem, precipitableWaterItem)
		station := sensorsTable[minItem.ID]

		currentObs := types.Observation{
			ObsTimeUtc:  minItem.At,
			StationID:   station.ID,
			StationName: station.Name,
			Lat:         station.Lat,
			Lon:         station.Lng,
			HumidityAvg: types.NaN(),
			WinddirAvg:  types.NaN(),
			Elevation:   station.Elevation,
			Metric: types.ObservationMetric{
				//DewptAvg:     types.NaN(),
				PrecipTotal:  types.NaN(),
				Pressure:     types.NaN(),
				TempAvg:      types.NaN(),
				WindspeedAvg: types.NaN(),
			},
		}

		if relativeHumidityItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(relativeHumidityItem.At) {
			currentObs.HumidityAvg = relativeHumidityItem.SensorValue()
			relativeHumidityIdx++
		}

		if temperatureItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(temperatureItem.At) {
			currentObs.Metric.TempAvg = temperatureItem.SensorValue()
			temperatureIdx++
		}

		if windDirectionItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(windDirectionItem.At) {
			currentObs.WinddirAvg = windDirectionItem.SensorValue()
			windDirectionIdx++
		}

		if windSpeedItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(windSpeedItem.At) {

			var value types.Value

			wsSensor := sensorsTable[windSpeedItem.ID]

			if wsSensor.MU == "Km/h" {
				// convert into m/s
				value = 0.277778 * windSpeedItem.SensorValue()
			} else if wsSensor.MU == "m/s" {
				value = windSpeedItem.SensorValue()
			} else {
				return nil, fmt.Errorf("unknown measure for wind speed in sensor %s: %s", windSpeedItem.ID, wsSensor.MU)
			}

			currentObs.Metric.WindspeedAvg = value

			windSpeedIdx++

		}

		if precipitableWaterItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(precipitableWaterItem.At) {
			currentObs.Metric.PrecipTotal = precipitableWaterItem.SensorValue()
			precipitableWaterIdx++
		}

		if pressureItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(pressureItem.At) {
			currentObs.Metric.Pressure = pressureItem.SensorValue()
			pressureIdx++
		} else {

			currentObs.Metric.Pressure = standardAtmosphere(station.Elevation)
		}

		/*
			// formula for dewpoint calculation must be applied with
			// temperature stil in celsius
			currentObs.CalculateDewpoint()
		*/

		// convert temperatures from °celsius to °kelvin
		//currentObs.Metric.DewptAvg += 273.15
		currentObs.Metric.TempAvg += 273.15

		// convert pression from hPa to Pa
		currentObs.Metric.Pressure *= 100

		results = append(results, currentObs)

	}

	return results, nil
}

type standardPressure struct {
	altMin, altMax           float64
	pressureMin, pressureMax float64
}

var standardValues = []standardPressure{
	{0, 1000, 101325, 89876},
	{1000, 5000, 89876, 54048},
	{5000, 10000, 54048, 26500},
	{10000, 15000, 26500, 12111},
	{15000, 20000, 12111, 5469},
	{20000, 25000, 5469, 2549},
	{25000, math.NaN(), 2549, math.NaN()},
}

func standardAtmosphere(elevation float64) types.Value {
	var level standardPressure
	for _, level = range standardValues {
		if level.altMin <= elevation && (math.IsNaN(level.altMax) || level.altMax > elevation) {
			break
		}
	}
	x0, x1 := level.altMin, level.altMax
	y0, y1 := level.pressureMin, level.pressureMax

	result := y0 + (elevation-x0)*(y1-y0)/(x1-x0)

	return types.Value(result / 100)
}

func openSensorsMap(dataPath string, domain types.Domain, sensorClass string) (map[string]sensorAnag, error) {
	sensorsTable := map[string]sensorAnag{}
	//fmt.Println("openSensorsMap", sensorClass, domain)

	err := fillSensorsMap(dataPath, domain, sensorClass, sensorsTable)
	if err != nil {
		return nil, err
	}

	return sensorsTable, nil
}

func fillSensorsMap(dataPath string, domain types.Domain, sensorClass string, sensorsTable map[string]sensorAnag) error {
	//fmt.Printf("fillSensorsMap %s\n", sensorClass)

	sensorsAnag := []sensorAnag{}
	///*testutil.FixtureDir(".."),*/ "../data"
	sensorsAnagContent, err := ioutil.ReadFile(path.Join(dataPath, sensorClass+"-registry.json"))
	if err != nil {
		return err
	}

	err = json.Unmarshal(sensorsAnagContent, &sensorsAnag)
	if err != nil {
		return err
	}

	for _, sensor := range sensorsAnag {
		if sensor.Lat >= domain.MinLat && sensor.Lat <= domain.MaxLat &&
			sensor.Lng >= domain.MinLon && sensor.Lng <= domain.MaxLon {
			sensor.Elevation = elevations.GetFromCoord(sensor.Lat, sensor.Lng)
			if _, exists := sensorsTable[sensor.ID]; exists {
				return fmt.Errorf("sensor exists with id %s", sensor.ID)
			}
			//fmt.Printf("%s sensor %s\n", sensorClass, sensor.ID)
			sensorsTable[sensor.ID] = sensor

		}
	}

	return nil
}

func openCompleteSensorsMap(dataPath string, domain types.Domain) (map[string]sensorAnag, error) {
	sensorsTable := map[string]sensorAnag{}

	//fmt.Println("openCompleteSensorsMap", domain)

	err := fillSensorsMap(dataPath, domain, "IGROMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap(dataPath, domain, "TERMOMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap(dataPath, domain, "DIREZIONEVENTO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap(dataPath, domain, "ANEMOMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap(dataPath, domain, "PLUVIOMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	err = fillSensorsMap(dataPath, domain, "BAROMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}

	return sensorsTable, nil
}
