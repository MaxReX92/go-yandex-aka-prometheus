package parser

import (
	"fmt"
	"strconv"
)

func ToFloat64(str string) (float64, error) {
	return strconv.ParseFloat(str, 64)
}

func ToInt64(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}

func FloatToString(num float64) string {
	return fmt.Sprintf("%g", num)
}

func IntToString(num int64) string {
	return strconv.FormatInt(num, 10)
}
