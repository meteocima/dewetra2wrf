package elevations

import (
	"fmt"
	"os"
	"testing"
)

func TestConvertToAscii(t *testing.T) {
	alt := GetFromCoord(45.589854, 1.7522)
	fmt.Fprint(os.Stderr, alt)
}
