package utils

import (
	"math"
	"strconv"
	"strings"

	"github.com/by46/whalefs/constant"
)

func Float64Equal(x, y float64) bool {
	return math.Abs(x-y) <= constant.Fraction
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

func ToInt32(value string) int32 {
	n, _ := strconv.ParseInt(value, 10, 32)
	return int32(n)
}

func ToBool(value string) bool {
	value = strings.ToLower(value)
	return value == constant.LiteralTrue
}
