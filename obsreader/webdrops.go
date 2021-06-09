package obsreader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"path"
	"path/filepath"
	"sort"
	"time"

	"github.com/meteocima/dewetra2wrf/elevations"
	"github.com/meteocima/dewetra2wrf/types"
)

// WebdropsObsReader is a struct that implements ObsReader
// and that reads observations from JSON files as downloaded
// from "webdrops" CIMA service.
type WebdropsObsReader struct{}

// ReadAll implements ObsReader for WebdropsObsReader
func (r WebdropsObsReader) ReadAll(dataPath string, domain types.Domain, date time.Time) ([]types.Observation, error) {
	/*
		relativeHumidity, err := readRelativeHumidity(dataPath, domain, date)
		if err != nil {
			return nil, err
		}
	*/
	temperature, err := readTemperature(dataPath, domain, date)
	if err != nil {
		return nil, err
	}
	/*
		windDirection, err := readWindDirection(dataPath, domain, date)
		if err != nil {
			return nil, err
		}

		windSpeed, err := readWindSpeed(dataPath, domain, date)
		if err != nil {
			return nil, err
		}

		precipitableWater, err := readPrecipitableWater(dataPath, domain, date)
		if err != nil {
			return nil, err
		}

		pressure, err := readPressure(dataPath, domain, date)
		if err != nil {
			return nil, err
		}
	*/
	return mergeObservations(dataPath, domain /*pressure, relativeHumidity, */, temperature /*, windDirection, windSpeed, precipitableWater*/)
}

/*
func readRelativeHumidity(dataPath string, domain types.Domain, date time.Time) ([]types.Result, error) {
	return readDewetraSensor(dataPath, domain, "IGROMETRO", date)
}
*/
func readTemperature(dataPath string, domain types.Domain, date time.Time) ([]types.Result, error) {
	return readDewetraSensor(dataPath, domain, "TERMOMETRO", date)
}

/*
func readWindDirection(dataPath string, domain types.Domain, date time.Time) ([]types.Result, error) {
	return readDewetraSensor(dataPath, domain, "DIREZIONEVENTO", date)
}

func readWindSpeed(dataPath string, domain types.Domain, date time.Time) ([]types.Result, error) {
	return readDewetraSensor(dataPath, domain, "ANEMOMETRO", date)
}

func readPrecipitableWater(dataPath string, domain types.Domain, date time.Time) ([]types.Result, error) {
	return readDewetraSensor(dataPath, domain, "PLUVIOMETRO", date)
}

func readPressure(dataPath string, domain types.Domain, date time.Time) ([]types.Result, error) {
	return readDewetraSensor(dataPath, domain, "BAROMETRO", date)
}
*/
func readDewetraSensor(dataPath string, domain types.Domain, sensorClass string, date time.Time) ([]types.Result, error) {

	content, err := ioutil.ReadFile(filepath.Join(dataPath, sensorClass+".json"))
	if err != nil {
		return nil, err
	}
	sensorObservations := []types.Result{}
	observationLess := func(i, j int) bool {
		if sensorObservations[i].SortKey == sensorObservations[j].SortKey {
			return sensorObservations[i].At.Unix() < sensorObservations[j].At.Unix()
		}
		return sensorObservations[i].SortKey < sensorObservations[j].SortKey
	}

	data := []sensorData{}

	err = json.Unmarshal(content, &data)
	if err != nil {
		return nil, err
	}

	sensorsTable, err := openSensorsMap(dataPath, domain, sensorClass)
	if err != nil {
		return nil, err
	}

	betterTimed := map[string]types.Result{}

	for _, sens := range data {
		sensAnag, ok := sensorsTable[sens.SensorID]
		if !ok {
			continue
		}
		for idx, dateS := range sens.Timeline {
			at, err := time.Parse(time.RFC3339, dateS)
			if err != nil {
				return nil, err
			}

			sortKey := fmt.Sprintf("%s:%05f:%05f", sensAnag.Name, sensAnag.Lat, sensAnag.Lng)
			//fmt.Println(sortKey)

			betterTimedObs, ok := betterTimed[sortKey]

			dateS := date.Format("20060102 15:04")
			atS := at.Format("20060102 15:04")
			betterS := betterTimedObs.At.Format("20060102 15:04")

			_, _, _ = dateS, atS, betterS

			if !ok || math.Abs(at.Sub(date).Minutes()) < math.Abs(betterTimedObs.At.Sub(date).Minutes()) {
				sensorResult := types.Result{
					At:      at,
					Value:   sens.Values[idx],
					SortKey: sortKey,
					ID:      sens.SensorID,
				}
				betterTimed[sortKey] = sensorResult
				//sensorObservations = append(sensorObservations, sensorResult)
			}
		}
	}

	for _, sens := range betterTimed {
		sensorObservations = append(sensorObservations, sens)
	}

	sort.SliceStable(sensorObservations, observationLess)

	return sensorObservations, nil
}

// mergeObservations is
func mergeObservations(dataPath string, domain types.Domain /*, pressure, relativeHumidity,*/, temperature /*, windDirection, windSpeed, precipitableWater*/ []types.Result) ([]types.Observation, error) {
	//pressureIdx := 0
	//relativeHumidityIdx := 0
	temperatureIdx := 0
	//windDirectionIdx := 0
	//windSpeedIdx := 0
	//precipitableWaterIdx := 0

	results := []types.Observation{}

	sensorsTable, err := openCompleteSensorsMap(dataPath, domain)
	if err != nil {
		return nil, err
	}

	for {
		/*
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
		*/
		var temperatureItem types.Result
		if len(temperature) > temperatureIdx {
			temperatureItem = temperature[temperatureIdx]
		} else {
			temperatureItem.SortKey = "zzzzzzzzzzzz"
		}
		/*
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
		*/
		if temperatureItem.SortKey == "zzzzzzzzzzzz" {
			break
		}

		minItem := temperatureItem // minObservation(pressureItem, relativeHumidityItem, temperatureItem, windDirectionItem, windSpeedItem, precipitableWaterItem)
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
		/*
			if relativeHumidityItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(relativeHumidityItem.At) {
				currentObs.HumidityAvg = relativeHumidityItem.SensorValue()
				relativeHumidityIdx++
			}
		*/
		if temperatureItem.SortKey == currentObs.SortKey() && currentObs.ObsTimeUtc.Equal(temperatureItem.At) {
			currentObs.Metric.TempAvg = temperatureItem.SensorValue()
			temperatureIdx++
		}
		/*
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
		*/

		/*
			// formula for dewpoint calculation must be applied with
			// temperature stil in celsius
			currentObs.CalculateDewpoint()
		*/

		// convert temperatures from °celsius to °kelvin
		//currentObs.Metric.DewptAvg += 273.15
		currentObs.Metric.TempAvg += 273.15

		// convert pression from hPa to Pa
		//currentObs.Metric.Pressure *= 100

		results = append(results, currentObs)

	}

	return results, nil
}

/*
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

/*
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
*/
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
	/*
		err := fillSensorsMap(dataPath, domain, "IGROMETRO", sensorsTable)
		if err != nil {
			return nil, err
		}
	*/
	err := fillSensorsMap(dataPath, domain, "TERMOMETRO", sensorsTable)
	if err != nil {
		return nil, err
	}
	/*
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
	*/
	return sensorsTable, nil
}

/*

> * wind speed - ANEMOMETRO
> * wind direction - DIREZIONEVENTO
> * dewpoint temperature - Non esiste, puoi calcolarla
> * temperature - TERMOMETRO
> * relative humidity - IGROMETRO
> * precipitable water - PLUVIOMETRO

*/
