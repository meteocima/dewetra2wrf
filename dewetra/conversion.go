package dewetra

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// QC is
const QC = 0

// ERROR is
const ERROR = 0.10

func skipLines(reader *bufio.Reader, n int) {
	for i := 0; i < n; i++ {
		_, err := reader.ReadString('\n')
		if err == io.EOF {
			return
		}
		if err != nil {
			panic(err)
		}
	}
}

func str(s string, len int) string {
	strFmt := fmt.Sprintf("%%%ds", len)
	return fmt.Sprintf(strFmt, s)
}

func integer(i int, len int) string {
	return str(fmt.Sprintf("%d", i), len)
}

func num(f SensorData, len float64) string {
	strFmt := fmt.Sprintf("%% %sf", strconv.FormatFloat(float64(len), 'f', -1, 64))
	return fmt.Sprintf(strFmt, f)
}

func space(n int) string {
	return strings.Repeat(" ", n)
}

func date(dt time.Time) string {
	return dt.Format("2006-01-02_15:04:05")
}

func dataQCError(data string) string {
	return data +
		integer(QC, 4) +
		num(ERROR, 7.2)
}

func printObservation(obs Observation) {
	name := "SAME"
	elevation := 0.0
	fmt.Println(
		str("FM-12 SYNOP", 12),
		" ",
		date(obs.ObsTimeUtc),
		" ",
		str(name, 40),
		" ",
		integer(1, 6),
		" ",
		num(SensorData(obs.Lat), 12.3),
		space(11),
		num(SensorData(obs.Lon), 12.3),
		space(11),
		num(SensorData(elevation), 12.3),
		space(11),
		space(6),
		str(obs.StationID, 40),
	)

	surfaceLevelPressure := SensorData(0.0)
	fmt.Println(
		dataQCError(num(surfaceLevelPressure, 12.3)),
		dataQCError(num(obs.Metric.PrecipTotal, 12.3)),
	)

	fmt.Println(
		dataQCError(num(obs.Metric.Pressure, 12.3)),
		dataQCError(num(obs.Metric.WindspeedAvg, 12.3)),
		dataQCError(num(obs.WinddirAvg, 12.3)),
		dataQCError(num(SensorData(elevation), 12.3)),
		dataQCError(num(obs.Metric.TempAvg, 12.3)),
		dataQCError(num(obs.Metric.DewptAvg, 12.3)),
		dataQCError(num(obs.HumidityAvg, 12.3)),
	)
}
