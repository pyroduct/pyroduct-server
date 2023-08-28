package state

import (
	"fmt"
	"strings"
	"time"
	"unicode"
)

type TimeSlice struct {
	Count uint32
	Start time.Time
	End   time.Time
	Next  *TimeSlice
}

type UsagePeriod struct {
	earliest        *TimeSlice
	latest          *TimeSlice
	Name            string
	TimeUnit        string
	timeUnitEnum    TimeUnit
	UnitMultiple    int
	Granularity     string
	granularityEnum TimeUnit
	noCullOnAccess  bool
	ClockAlign      bool
	Limit           uint64
	resetClockTime  time.Time
	count           uint64
	timeFunc        func() time.Time
	duration        time.Duration
}

func (up *UsagePeriod) Allowed() (bool, int) {

	now := up.timeFunc()

	up.Trim(now)

	if up.latest == nil {

		up.latest = up.buildSlice(now)
		up.earliest = up.latest

		if up.ClockAlign {
			_, up.resetClockTime = up.calculateAlignedStartEndTimes(now, up.timeUnitEnum)
		}

	} else {
		up.updateLatest(now)
	}

	if up.count >= up.Limit {
		return false, 0
	} else {
		up.latest.Count += 1
		up.count += 1

		return true, 0
	}

}

func (up *UsagePeriod) updateLatest(now time.Time) {

	if now.After(up.latest.End) {
		ns := up.buildSlice(now)

		up.latest.Next = ns
		up.latest = ns
	}

}

func (up *UsagePeriod) buildSlice(now time.Time) *TimeSlice {
	ts := new(TimeSlice)

	ts.Start, ts.End = up.sliceStartEnd(now)

	return ts
}

func (up *UsagePeriod) Trim(now time.Time) {

	if up.earliest == nil {
		return
	}

	if up.ClockAlign && !up.resetClockTime.IsZero() && now.After(up.resetClockTime) {
		up.Reset()
		return
	}

}

func (up *UsagePeriod) Reset() {
	up.earliest = nil
	up.latest = nil
	up.count = 0

	return
}

func (up *UsagePeriod) sliceStartEnd(now time.Time) (time.Time, time.Time) {

	if up.ClockAlign {
		s, e := up.calculateAlignedStartEndTimes(now, up.granularityEnum)

		return s, e
	}

	return now, now
}

func (up *UsagePeriod) TotalSlices() int {
	if up.earliest == nil {
		return 0
	}

	return up.sliceCount(up.earliest, 0)
}

func (up *UsagePeriod) sliceCount(ts *TimeSlice, count int) int {

	count += 1

	if ts.Next == nil {
		return count
	} else {
		return up.sliceCount(ts.Next, count)
	}

}

func (up *UsagePeriod) calculateAlignedStartEndTimes(t time.Time, tu TimeUnit) (time.Time, time.Time) {

	//startTemplate := "2006-01-02T15:04:05.999999999Z"
	startTemplate := "%d-%02d-%02dT%02d:%02d:%02d.000000000Z"

	day, month, year := t.Day(), t.Month(), t.Year()
	var hour, minutes, seconds int

	if tu < DAY {
		hour = t.Hour()
	}

	if tu < HOUR {
		minutes = t.Minute()
	}

	if tu < MINUTE {
		seconds = t.Second()
	}

	ts := fmt.Sprintf(startTemplate, year, month, day, hour, minutes, seconds)
	start, _ := time.Parse(time.RFC3339Nano, ts)

	end := start.Add(unitToDuration(tu, time.Duration(-1)))

	return start, end
}

//type timeFunc func() time.Time

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

func unitToDuration(unit TimeUnit, offset time.Duration) time.Duration {
	switch unit {
	case DAY:
		return (time.Hour * 24) + offset
	case SECOND:
		return time.Second + offset
	case MINUTE:
		return time.Minute + offset
	case HOUR:
		return time.Hour + offset
	}

	return 0
}

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

	return -1, fmt.Errorf("no such supported TimeUnit \"%s\"", label)
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

func (up *UsagePeriod) Initialise() error {

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
	} else {
		up.Granularity = up.TimeUnit
		up.granularityEnum = up.timeUnitEnum
	}

	if up.granularityEnum > up.timeUnitEnum {
		return fmt.Errorf("value for field Granularity %s is greater than the value of TimeUnit %s", up.Granularity, up.TimeUnit)
	}

	if up.UnitMultiple < 0 {
		return fmt.Errorf("UnitMultiple must be unset or >= 1")
	} else if up.UnitMultiple == 0 {
		up.UnitMultiple = 1
	}

	up.duration = unitToDuration(up.timeUnitEnum, 0) * time.Duration(up.UnitMultiple)

	if up.Limit <= 0 {
		return fmt.Errorf("value for field Limit %d must be greater than zero", up.Limit)
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
