package utils

import (
	"log"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

func HandlePanic() {
	if r := recover(); r != nil {
		log.Println("RECOVER", r)
		debug.PrintStack()
	}
}

var sep = "_"

func TimeToString(time time.Time, includeNanoSec bool) string {
	arr := make([]string, 0)
	arr = append(arr, strconv.Itoa(time.Year()))
	arr = append(arr, fixZero(int(time.Month())))
	arr = append(arr, fixZero(time.Day()))
	arr = append(arr, fixZero(time.Hour()))
	arr = append(arr, fixZero(time.Minute()))
	arr = append(arr, fixZero(time.Second()))
	if includeNanoSec {
		arr = append(arr, fixZero(time.Nanosecond()))
	}

	return strings.Join(arr, sep)
}

func fixZero(val int) string {
	if val < 10 {
		return "0" + strconv.Itoa(val)
	}
	return strconv.Itoa(val)
}
