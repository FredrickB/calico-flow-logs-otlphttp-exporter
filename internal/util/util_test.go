package util

import (
	"testing"
	"time"
)

func TestShouldParseSecondsCorrectly(t *testing.T) {
	milliSecondStringValue := "5"
	defaultValue := 10 * time.Millisecond
	expected := 5 * time.Millisecond
	actual := ParseMilliSecondsStringValue(milliSecondStringValue, defaultValue)

	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}

func TestInvalidSecondsReturnsDefaultValue(t *testing.T) {
	milliSecondStringValue := "invalid milliseconds"
	defaultValue := 10 * time.Millisecond
	expected := defaultValue
	actual := ParseMilliSecondsStringValue(milliSecondStringValue, defaultValue)

	if actual != expected {
		t.Errorf("expected: %s, actual: %s", expected, actual)
	}
}
