package elevations

import (
	"os"
	"path"
)

type ElevationsFile struct {
	xs, ys []float64
	zs     []int32
}

var elev *ElevationsFile = openElevationsFile()

func openElevationsFile() *ElevationsFile {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	orog := path.Join(home, ".dewetra2wrf", "orog.nc")
	f := File{}
	f.Open(orog)
	defer f.Close()
	x := f.Var("x")
	y := f.Var("y")
	z := f.Var("z")
	e := &ElevationsFile{
		xs: x.ValuesFloat64(),
		ys: y.ValuesFloat64(),
		zs: z.ValuesInt32(),
	}
	if f.Error() != nil {
		panic(err)
	}

	return e

}

func GetFromCoord(lat, lon float64) float64 {
	ypos := int((0.5 + lat/180) * float64(len(elev.ys)))
	xpos := int((0.5 + lon/360) * float64(len(elev.xs)))

	val := float64(elev.zs[xpos+ypos*len(elev.xs)])
	return val
}
