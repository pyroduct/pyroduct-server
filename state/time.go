package state

import (
	"fmt"
	"strings"
	"unicode"
)

type TimeSlice struct {
	Count uint32
	Next  *TimeSlice
}

type UsagePeriod struct {
	first           *TimeSlice
	Last            *TimeSlice
	Name            string
	TimeUnit        string
	timeUnitEnum    TimeUnit
	UnitMultiple    int
	Granularity     string
	granularityEnum TimeUnit
	noCullOnAccess  bool
	ClockAlign      bool
	count           int64
}

type TimeUnit int

const (
	SECOND = 100
	MINUTE = 200
	HOUR   = 300
	DAY    = 400
)

const (
	SECOND_LABEL = "SECOND"
	MINUTE_LABEL = "MINUTE"
	HOUR_LABEL   = "HOUR"
	DAY_LABEL    = "DAY"
)

func labelToUnit(label string) (TimeUnit, error) {
	switch label {
	case DAY_LABEL:
		return DAY, nil
	case SECOND_LABEL:
		return SECOND, nil
	case MINUTE_LABEL:
		return MINUTE, nil
	case HOUR_LABEL:
		return HOUR, nil
	}

	return -1, fmt.Errorf("No such supported TimeUnit \"%s\"", label)
}

func unitToLabel(unit TimeUnit) (string, error) {
	switch unit {
	case DAY:
		return DAY_LABEL, nil
	case SECOND:
		return SECOND_LABEL, nil
	case MINUTE:
		return MINUTE_LABEL, nil
	case HOUR:
		return HOUR_LABEL, nil
	}

	return "", fmt.Errorf("TimeUnit of value %d cannot be mapped to a label", unit)
}

func parseAndValidateUsagePeriod(up *UsagePeriod) error {

	var err error

	if !isSet(up.Name) {
		return fmt.Errorf("you must set the Name field")
	} else {
		up.Name = strings.TrimSpace(up.Name)
	}

	if containsWhiteSpace(up.Name) {
		return fmt.Errorf("the field Name must not contain whitespace")
	}

	if !isSet(up.TimeUnit) {
		return fmt.Errorf("you must set the TimeUnit field")
	}

	if up.timeUnitEnum, err = labelToUnit(up.TimeUnit); err != nil {
		return err
	}

	if isSet(up.Granularity) {
		if up.granularityEnum, err = labelToUnit(up.Granularity); err != nil {
			return err
		}
	}

	if up.granularityEnum > up.timeUnitEnum {
		return fmt.Errorf("value for field Granularity %s is greater than the value of TimeUnit %s", up.Granularity, up.TimeUnit)
	}

	if up.UnitMultiple < 0 {
		return fmt.Errorf("UnitMultiple must be unset or >= 1")
	} else if up.UnitMultiple == 0 {
		up.UnitMultiple = 1
	}

	return nil
}

func isSet(s string) bool {
	return strings.TrimSpace(s) != ""
}

func containsWhiteSpace(s string) bool {
	for _, r := range s {
		if unicode.IsSpace(r) {
			return true
		}
	}

	return false
}
