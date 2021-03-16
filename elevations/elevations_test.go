package elevations

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertToAscii(t *testing.T) {
	alt := GetFromCoord(44.4895755, 8.9287799)
	assert.Equal(t, 236.0, alt)
}
