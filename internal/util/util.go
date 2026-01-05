package util

import (
	"log"
	"strconv"
	"time"
)

func LogEnvironmentVariable(variable, value string) {
	log.Printf("%s set to %s", variable, value)
}

func ParseMilliSecondsStringValue(milliSecondsAsString string, defaultValue time.Duration) time.Duration {
	secondsParsed, err := strconv.Atoi(milliSecondsAsString)
	if err != nil {
		return defaultValue
	} else {
		return time.Duration(secondsParsed) * time.Millisecond
	}
}
