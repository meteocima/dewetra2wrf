package elevations

import (
	"os"
	"path"
	"sync"
)

type ElevationsFile struct {
	xs, ys []float64
	zs     []int32
}

var elev *ElevationsFile
var elevLock sync.Mutex

func OpenElevationsFile(dirname string) (*ElevationsFile, error) {
	home, _ := os.UserHomeDir()
	orog := path.Join(home, ".dewetra2wrf", "orog.nc")
	elevLock.Lock()
	defer elevLock.Unlock()
	var err error
	if elev != nil {
		f := File{}
		f.Open(orog)
		defer f.Close()
		x := f.Var("x")
		y := f.Var("y")
		z := f.Var("z")
		elev = &ElevationsFile{
			xs: x.ValuesFloat64(),
			ys: y.ValuesFloat64(),
			zs: z.ValuesInt32(),
		}
		err = f.Error()
	}

	return elev, err

}

func (file *ElevationsFile) GetElevation(lat, lon float64) float64 {
	ypos := int((0.5 + lat/180) * float64(len(file.ys)))
	xpos := int((0.5 + lon/360) * float64(len(file.xs)))

	val := float64(file.zs[xpos+ypos*len(file.xs)])
	return val
}
