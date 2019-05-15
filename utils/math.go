package utils

import (
	"math"
	"strconv"
	"strings"
)

const (
	Fraction = 0.000001
)

func Float64Equal(x, y float64) bool {
	return math.Abs(math.Dim(x, y)) <= Fraction
}

func RatioEval(ratio string) *float64 {
	segments := strings.SplitN(ratio, ":", 2)
	if len(segments) != 2 {
		return nil
	}
	x, err := strconv.ParseFloat(segments[0], 64)
	if err != nil {
		return nil
	}
	y, err := strconv.ParseFloat(segments[1], 64)
	if err != nil {
		return nil
	}
	if y == 0.0 {
		return nil
	}
	r := x / y
	return &r
}
