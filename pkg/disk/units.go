package disk

import (
	"github.com/alecthomas/units"
	"github.com/pkg/errors"
	"strconv"
)

var (
	// See https://en.wikipedia.org/wiki/Binary_prefix
	shortBinaryUnitMap = units.MakeUnitMap("", "c", 1024)
	binaryUnitMap      = units.MakeUnitMap("iB", "c", 1024)
	decimalUnitMap     = units.MakeUnitMap("B", "c", 1000)
)

// ParseUnit parse a digit with unit such as "K" , "KiB", "KB", "c", "MiB", "MB", "M".
// If input string is a digit without unit ,
// it will be regarded as a digit with unit M(1024*1024 bytes).
func ParseUnit(s string) (uint64, error) {
	if _, err := strconv.Atoi(s); err == nil {
		s += "M"
	}
	if n, err := units.ParseUnit(s, shortBinaryUnitMap); err == nil {
		return uint64(n), nil
	}

	if n, err := units.ParseUnit(s, binaryUnitMap); err == nil {
		return uint64(n), nil
	}

	if n, err := units.ParseUnit(s, decimalUnitMap); err == nil {
		return uint64(n), nil
	}
	return 0, errors.Errorf("units: unknown unit %s", s)
}
