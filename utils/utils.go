package utils

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

func HandlePanic() {
	if r := recover(); r != nil {
		fmt.Println("RECOVER", r)
		debug.PrintStack()
	}
}

var sep = "_"

func TimeToString(time time.Time, includeNanoSec bool) string {
	arr := make([]string, 0)
	arr = append(arr, strconv.Itoa(time.Year()))
	arr = append(arr, strconv.Itoa(int(time.Month())))
	arr = append(arr, strconv.Itoa(time.Day()))
	arr = append(arr, strconv.Itoa(time.Hour()))
	arr = append(arr, strconv.Itoa(time.Minute()))
	arr = append(arr, strconv.Itoa(time.Second()))
	if includeNanoSec {
		arr = append(arr, strconv.Itoa(time.Nanosecond()))
	}

	return strings.Join(arr, sep)
}
