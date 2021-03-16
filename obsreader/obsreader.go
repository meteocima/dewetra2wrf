package obsreader

import (
	"time"

	"github.com/meteocima/dewetra2wrf/types"
)

// ObsReader is implemented by types that
// are ables to read `types.Observation`
type ObsReader interface {
	// ReadAll returns a slice of types.Observation read
	// from path argument, filtered by `domain` and
	// `date`
	ReadAll(path string, domain types.Domain, date time.Time) ([]types.Observation, error)
}
