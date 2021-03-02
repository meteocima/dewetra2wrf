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

	"github.com/meteocima/dewetra2wrf/obsreader/elevations"
	"github.com/meteocima/dewetra2wrf/obsreader/wunderground"
	"github.com/meteocima/dewetra2wrf/sensor"
)

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

type sensorReqBody struct {
	SensorClass string   `json:"sensorClass"`
	From        string   `json:"from"`
	To          string   `json:"to"`
	Ids         []string `json:"ids"`
}

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

func observationIsLess(this, that sensor.Result) bool {
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

func minObservation(results ...sensor.Result) sensor.Result {
	min := sensor.Result{SortKey: "zzzzzzzzzzzz"}
	for _, result := range results {
		if observationIsLess(result, min) {
			min = result
		}
	}
	return min
}

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

// AllSensorsWund is
func AllSensorsWund(dataPath string, domain sensor.Domain, date time.Time) ([]sensor.Observation, error) {
	return wunderground.Read(dataPath, domain, date)
}

// AllSensorsWundHistory is
func AllSensorsWundHistory(dataPath string, domain sensor.Domain, date time.Time) ([]sensor.Observation, error) {
	return wunderground.ReadHistory(dataPath, domain, date)
}

// AllSensors is
func AllSensors(dataPath string, domain sensor.Domain, date time.Time) ([]sensor.Observation, error) {

	relativeHumidity, err := readRelativeHumidity(dataPath, domain, date)
	if err != nil {
		return nil, err
	}

	temperature, err := readTemperature(dataPath, domain, date)
	if err != nil {
		return nil, err
	}

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

	return MergeObservations(dataPath, domain, pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater)
}

// MergeObservations is
func MergeObservations(dataPath string, domain sensor.Domain, pressure, relativeHumidity, temperature, windDirection, windSpeed, precipitableWater []sensor.Result) ([]sensor.Observation, error) {
	pressureIdx := 0
	relativeHumidityIdx := 0
	temperatureIdx := 0
	windDirectionIdx := 0
	windSpeedIdx := 0
	precipitableWaterIdx := 0

	results := []sensor.Observation{}

	sensorsTable, err := openCompleteSensorsMap(dataPath, domain)
	if err != nil {
		return nil, err
	}

	for {
		var pressureItem sensor.Result
		if len(pressure) > pressureIdx {
			pressureItem = pressure[pressureIdx]
		} else {
			pressureItem.SortKey = "zzzzzzzzzzzz"
		}

		var relativeHumidityItem sensor.Result
		if len(relativeHumidity) > relativeHumidityIdx {
			relativeHumidityItem = relativeHumidity[relativeHumidityIdx]
		} else {
			relativeHumidityItem.SortKey = "zzzzzzzzzzzz"
		}

		var temperatureItem sensor.Result
		if len(temperature) > temperatureIdx {
			temperatureItem = temperature[temperatureIdx]
		} else {
			temperatureItem.SortKey = "zzzzzzzzzzzz"
		}

		var windDirectionItem sensor.Result
		if len(windDirection) > windDirectionIdx {
			windDirectionItem = windDirection[windDirectionIdx]
		} else {
			windDirectionItem.SortKey = "zzzzzzzzzzzz"
		}

		var windSpeedItem sensor.Result
		if len(windSpeed) > windSpeedIdx {
			windSpeedItem = windSpeed[windSpeedIdx]
		} else {
			windSpeedItem.SortKey = "zzzzzzzzzzzz"
		}

		var precipitableWaterItem sensor.Result
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

		currentObs := sensor.Observation{
			ObsTimeUtc:  minItem.At,
			StationID:   station.ID,
			StationName: station.Name,
			Lat:         station.Lat,
			Lon:         station.Lng,
			HumidityAvg: sensor.NaN(),
			WinddirAvg:  sensor.NaN(),
			Elevation:   station.Elevation,
			Metric: sensor.ObservationMetric{
				//DewptAvg:     sensor.NaN(),
				PrecipTotal:  sensor.NaN(),
				Pressure:     sensor.NaN(),
				TempAvg:      sensor.NaN(),
				WindspeedAvg: sensor.NaN(),
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

			value := sensor.NaN()

			wsSensor := sensorsTable[windSpeedItem.ID]

			if wsSensor.MU == "Km/h" {
				// convert into m/s
				value = 0.277778 * windSpeedItem.SensorValue()
			} else if wsSensor.MU == "m/s" {
				value = windSpeedItem.SensorValue()
			} else {
				return nil, fmt.Errorf("Unknown measure for wind speed in sensor %s: %s", windSpeedItem.ID, wsSensor.MU)
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

func standardAtmosphere(elevation float64) sensor.Value {
	var level standardPressure
	for _, level = range standardValues {
		if level.altMin <= elevation && (math.IsNaN(level.altMax) || level.altMax > elevation) {
			break
		}
	}
	x0, x1 := level.altMin, level.altMax
	y0, y1 := level.pressureMin, level.pressureMax

	result := y0 + (elevation-x0)*(y1-y0)/(x1-x0)

	return sensor.Value(result / 100)
}

func readRelativeHumidity(dataPath string, domain sensor.Domain, date time.Time) ([]sensor.Result, error) {
	return readDewetraSensor(dataPath, domain, "IGROMETRO", date)
}

func readTemperature(dataPath string, domain sensor.Domain, date time.Time) ([]sensor.Result, error) {
	return readDewetraSensor(dataPath, domain, "TERMOMETRO", date)
}

func readWindDirection(dataPath string, domain sensor.Domain, date time.Time) ([]sensor.Result, error) {
	return readDewetraSensor(dataPath, domain, "DIREZIONEVENTO", date)
}

func readWindSpeed(dataPath string, domain sensor.Domain, date time.Time) ([]sensor.Result, error) {
	return readDewetraSensor(dataPath, domain, "ANEMOMETRO", date)
}

func readPrecipitableWater(dataPath string, domain sensor.Domain, date time.Time) ([]sensor.Result, error) {
	return readDewetraSensor(dataPath, domain, "PLUVIOMETRO", date)
}

func readPressure(dataPath string, domain sensor.Domain, date time.Time) ([]sensor.Result, error) {
	return readDewetraSensor(dataPath, domain, "BAROMETRO", date)
}

func readDewetraSensor(dataPath string, domain sensor.Domain, sensorClass string, date time.Time) ([]sensor.Result, error) {

	content, err := ioutil.ReadFile(filepath.Join(dataPath, sensorClass+".json"))
	if err != nil {
		return nil, err
	}
	sensorObservations := []sensor.Result{}
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

	betterTimed := map[string]sensor.Result{}

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
				sensorResult := sensor.Result{
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

func openSensorsMap(dataPath string, domain sensor.Domain, sensorClass string) (map[string]sensorAnag, error) {
	sensorsTable := map[string]sensorAnag{}
	//fmt.Println("openSensorsMap", sensorClass, domain)

	err := fillSensorsMap(dataPath, domain, sensorClass, sensorsTable)
	if err != nil {
		return nil, err
	}

	return sensorsTable, nil
}

func fillSensorsMap(dataPath string, domain sensor.Domain, sensorClass string, sensorsTable map[string]sensorAnag) error {
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

	elevations, err := elevations.OpenElevationsFile(dataPath)
	if err != nil {
		return err
	}

	for _, sensor := range sensorsAnag {
		if sensor.Lat >= domain.MinLat && sensor.Lat <= domain.MaxLat &&
			sensor.Lng >= domain.MinLon && sensor.Lng <= domain.MaxLon {
			sensor.Elevation = elevations.GetElevation(sensor.Lat, sensor.Lng)
			if _, exists := sensorsTable[sensor.ID]; exists {
				return fmt.Errorf("Sensor exists with id %s", sensor.ID)
			}
			//fmt.Printf("%s sensor %s\n", sensorClass, sensor.ID)
			sensorsTable[sensor.ID] = sensor

		}
	}

	return nil
}

func openCompleteSensorsMap(dataPath string, domain sensor.Domain) (map[string]sensorAnag, error) {
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
