package parser

import (
	"fmt"
	"strconv"
)

// ToFloat64 parse strings to float64
func ToFloat64(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

// ToInt64 parse string to int64
func ToInt64(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

// FloatToString convert float64 to string.
func FloatToString(num float64) string {
	return fmt.Sprintf("%g", num)
}

// IntToString convert int64 to string.
func IntToString(num int64) string {
	return strconv.FormatInt(num, 10)
}
