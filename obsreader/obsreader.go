package obsreader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"

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
	StationName         string
	Lon, Lat, Elevation float64
}

func observationIsLess(this, that sensor.Result) bool {
	if this.SortKey == that.SortKey {
		return this.At.Unix() < that.At.Unix()
	}
	return this.SortKey < that.SortKey
}

func minObservation(results ...sensor.Result) sensor.Result {
	min := sensor.Result{SortKey: "zzzzzzzzzzzzzzzzzzzzzzzzz"}
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
	return bn.table[idI].StationName < bn.table[idJ].StationName
}

// Swap swaps the elements with indexes i and j.
func (bn byName) Swap(i, j int) {
	save := bn.ids[i]
	bn.ids[i] = bn.ids[j]
	bn.ids[j] = save
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
			pressureItem.SortKey = "zzzzzzzzzz"
		}

		var relativeHumidityItem sensor.Result
		if len(relativeHumidity) > relativeHumidityIdx {
			relativeHumidityItem = relativeHumidity[relativeHumidityIdx]
		} else {
			relativeHumidityItem.SortKey = "zzzzzzzzzz"
		}

		var temperatureItem sensor.Result
		if len(temperature) > temperatureIdx {
			temperatureItem = temperature[temperatureIdx]
		} else {
			temperatureItem.SortKey = "zzzzzzzzzz"
		}

		var windDirectionItem sensor.Result
		if len(windDirection) > windDirectionIdx {
			windDirectionItem = windDirection[windDirectionIdx]
		} else {
			windDirectionItem.SortKey = "zzzzzzzzzz"
		}

		var windSpeedItem sensor.Result
		if len(windSpeed) > windSpeedIdx {
			windSpeedItem = windSpeed[windSpeedIdx]
		} else {
			windSpeedItem.SortKey = "zzzzzzzzzz"
		}

		var precipitableWaterItem sensor.Result
		if len(precipitableWater) > precipitableWaterIdx {
			precipitableWaterItem = precipitableWater[precipitableWaterIdx]
		} else {
			precipitableWaterItem.SortKey = "zzzzzzzzzz"
		}

		if relativeHumidityItem.SortKey == "zzzzzzzzzz" &&
			temperatureItem.SortKey == "zzzzzzzzzz" &&
			windDirectionItem.SortKey == "zzzzzzzzzz" &&
			windSpeedItem.SortKey == "zzzzzzzzzz" &&
			precipitableWaterItem.SortKey == "zzzzzzzzzz" &&
			pressureItem.SortKey == "zzzzzzzzzz" {
			break
		}

		minItem := minObservation(pressureItem, relativeHumidityItem, temperatureItem, windDirectionItem, windSpeedItem, precipitableWaterItem)
		station := sensorsTable[minItem.ID]

		currentObs := sensor.Observation{
			ObsTimeUtc:  minItem.At,
			StationID:   station.ID,
			StationName: station.StationName,
			Lat:         station.Lat,
			Lon:         station.Lon,
			HumidityAvg: sensor.Value(math.NaN()),
			WinddirAvg:  sensor.Value(math.NaN()),
			Elevation:   station.Elevation,
			Metric: sensor.ObservationMetric{
				DewptAvg:     sensor.Value(math.NaN()),
				PrecipTotal:  sensor.Value(math.NaN()),
				Pressure:     sensor.Value(math.NaN()),
				TempAvg:      sensor.Value(math.NaN()),
				WindspeedAvg: sensor.Value(math.NaN()),
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
			currentObs.Metric.WindspeedAvg = windSpeedItem.SensorValue()
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

		// formula for dewpoint calculation must be applied with
		// temperature stil in celsius
		currentObs.CalculateDewpoint()

		// convert temperatures from °celsius to °kelvin
		currentObs.Metric.DewptAvg += 273.15
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
	standardPressure{0, 1000, 101325, 89876},
	standardPressure{1000, 5000, 89876, 54048},
	standardPressure{5000, 10000, 54048, 26500},
	standardPressure{10000, 15000, 26500, 12111},
	standardPressure{15000, 20000, 12111, 5469},
	standardPressure{20000, 25000, 5469, 2549},
	standardPressure{25000, math.NaN(), 2549, math.NaN()},
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
		for idx, dateS := range sens.Timeline {
			at, err := time.Parse("200601021504", dateS)
			if err != nil {
				return nil, err
			}

			//at = at.Add(time.Hour * 2)

			sensAnag := sensorsTable[sens.SensorID]
			sortKey := fmt.Sprintf("%s:%05f:%05f", sensAnag.StationName, sensAnag.Lat, sensAnag.Lon)

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

	err := fillSensorsMap(dataPath, domain, sensorClass, sensorsTable)
	if err != nil {
		return nil, err
	}

	return sensorsTable, nil
}

func getElevation(sensor sensorAnag) (float64, error) {
	url := fmt.Sprintf("https://api.airmap.com/elevation/v1/ele/?points=%6f,%6f", sensor.Lat, sensor.Lon)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("X-API-Key", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjcmVkZW50aWFsX2lkIjoiY3JlZGVudGlhbHx2YkJZOEs2dTAzRFdYZ1NaZDRvbkVodjJhMG9SIiwiYXBwbGljYXRpb25faWQiOiJhcHBsaWNhdGlvbnxEbzhFV1hKQzBOOEU4elVwcW45dkxUbFBYUE1SIiwib3JnYW5pemF0aW9uX2lkIjoiZGV2ZWxvcGVyfGE2Tk9LNUJ0S3ZBeHZkY2V3SzJrdmZ3ZUo5cUUiLCJpYXQiOjE1ODk3OTA2MTd9.dzqy2VbQtmHrf4sKwlb3S0PdLiqGS4ms4LFvKeWmMkY")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return 0, fmt.Errorf("HTTP response %d", res.StatusCode)
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	var data struct {
		Data []float64
	}

	err = json.Unmarshal(content, &data)
	if err != nil {
		return 0, err
	}

	return data.Data[0], nil
}

func readElevationsFromFile(dataPath string) (map[string]float64, error) {
	elevFile := path.Join(dataPath, "elevations.json")
	elevations := map[string]float64{}

	elevationsContent, err := ioutil.ReadFile(elevFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err == nil {
		err = json.Unmarshal(elevationsContent, &elevations)
		if err != nil {
			return nil, err
		}
	}

	return elevations, nil
}

func saveElevationsToFile(dataPath string, elevations map[string]float64) error {
	buff, err := json.MarshalIndent(elevations, " ", " ")
	if err != nil {
		return err
	}
	elevFile := path.Join(dataPath, "elevations.json")

	return ioutil.WriteFile(elevFile, buff, os.FileMode(0644))

}

func fillSensorsMap(dataPath string, domain sensor.Domain, sensorClass string, sensorsTable map[string]sensorAnag) error {
	sensorsAnag := []sensorAnag{}
	///*testutil.FixtureDir(".."),*/ "../data"
	sensorsAnagContent, err := ioutil.ReadFile(path.Join(dataPath, sensorClass+".json"))
	if err != nil {
		return err
	}

	err = json.Unmarshal(sensorsAnagContent, &sensorsAnag)
	if err != nil {
		return err
	}

	elevations, err := readElevationsFromFile(dataPath)
	if err != nil {
		return err
	}

	for n, sensor := range sensorsAnag {
		stKey := fmt.Sprintf("%f:%f", sensor.Lat, sensor.Lon)
		elevation, ok := elevations[stKey]
		if !ok {
			fmt.Println(n, stKey)
			elevation, err := getElevation(sensor)
			if err != nil {
				fmt.Println(err)
				continue
			}
			elevations[stKey] = elevation
			err = saveElevationsToFile(dataPath, elevations)
			if err != nil {
				return err
			}
		}

		sensor.Elevation = elevation
		if sensor.Lat >= domain.MinLat && sensor.Lat <= domain.MaxLat &&
			sensor.Lon >= domain.MinLon && sensor.Lon <= domain.MaxLon {
			sensorsTable[sensor.ID] = sensor

		}
	}

	return nil
}

func openCompleteSensorsMap(dataPath string, domain sensor.Domain) (map[string]sensorAnag, error) {
	sensorsTable := map[string]sensorAnag{}

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