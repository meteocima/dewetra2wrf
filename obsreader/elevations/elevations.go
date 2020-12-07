package elevations

type ElevationsFile struct {
	xs, ys []float64
	zs     []int32
}

func OpenElevationsFile(dirname string) (*ElevationsFile, error) {

	orog := "~/.dewetra2wrf/orog.nc"
	elev := &ElevationsFile{}
	f := File{}
	f.Open(orog)
	defer f.Close()
	x := f.Var("x")
	y := f.Var("y")
	z := f.Var("z")

	elev.xs = x.ValuesFloat64()
	elev.ys = y.ValuesFloat64()
	elev.zs = z.ValuesInt32()

	return elev, f.Error()

}

func (file *ElevationsFile) GetElevation(lat, lon float64) float64 {
	ypos := int((0.5 + lat/180) * float64(len(file.ys)))
	xpos := int((0.5 + lon/360) * float64(len(file.xs)))

	val := float64(file.zs[xpos+ypos*len(file.xs)])
	return val
}
