package state

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestLabelToTimeUnit(t *testing.T) {

	tu, err := labelToUnit("DAY")
	assert.Nil(t, err)
	assert.Equal(t, TimeUnit(DAY), tu)

	tu, err = labelToUnit("SECOND")
	assert.Nil(t, err)
	assert.Equal(t, TimeUnit(SECOND), tu)

	tu, err = labelToUnit("MINUTE")
	assert.Nil(t, err)
	assert.Equal(t, TimeUnit(MINUTE), tu)

	tu, err = labelToUnit("HOUR")
	assert.Nil(t, err)
	assert.Equal(t, TimeUnit(HOUR), tu)

	tu, err = labelToUnit("UNKNOWN")
	assert.NotNil(t, err)
	assert.Equal(t, TimeUnit(-1), tu)

}

func TestTimeUnitToLabel(t *testing.T) {

	tu, err := unitToLabel(DAY)
	assert.Nil(t, err)
	assert.Equal(t, "DAY", tu)

	tu, err = unitToLabel(SECOND)
	assert.Nil(t, err)
	assert.Equal(t, "SECOND", tu)

	tu, err = unitToLabel(MINUTE)
	assert.Nil(t, err)
	assert.Equal(t, "MINUTE", tu)

	tu, err = unitToLabel(HOUR)
	assert.Nil(t, err)
	assert.Equal(t, "HOUR", tu)

	tu, err = unitToLabel(TimeUnit(-10))
	assert.NotNil(t, err)
	assert.Equal(t, "", tu)

}

func TestValidUsagePeriodParse(t *testing.T) {

	up := new(UsagePeriod)

	up.Name = "test-period"
	up.TimeUnit = DAY_LABEL
	up.Limit = 10

	assert.Nil(t, up.Initialise())

	up.Granularity = HOUR_LABEL
	assert.Nil(t, up.Initialise())

	assert.Equal(t, TimeUnit(DAY), up.timeUnitEnum)
	assert.Equal(t, TimeUnit(HOUR), up.granularityEnum)
	assert.Equal(t, 1, up.UnitMultiple)

}

func TestNameValidation(t *testing.T) {

	up := new(UsagePeriod)

	up.Name = "test-period"
	up.TimeUnit = DAY_LABEL
	up.Limit = 10

	assert.Nil(t, up.Initialise())

	up.Name = "test period"
	assert.NotNil(t, up.Initialise())

	up.Name = "test\tperiod"
	assert.NotNil(t, up.Initialise())

	up.Name = "test\nperiod"
	assert.NotNil(t, up.Initialise())
}

func TestUnitGranularityRelationship(t *testing.T) {

	up := new(UsagePeriod)

	up.Name = "test-period"
	up.TimeUnit = DAY_LABEL
	up.Granularity = DAY_LABEL
	up.Limit = 10

	assert.Nil(t, up.Initialise())

	up.Granularity = HOUR_LABEL

	assert.Nil(t, up.Initialise())

	up.TimeUnit = SECOND_LABEL
	up.Granularity = MINUTE_LABEL

	assert.NotNil(t, up.Initialise())

}

func TestUnitInvalidUnitMultiple(t *testing.T) {

	up := new(UsagePeriod)

	up.Name = "test-period"
	up.TimeUnit = DAY_LABEL
	up.UnitMultiple = -1

	assert.NotNil(t, up.Initialise())
}

func TestMissingMandatory(t *testing.T) {

	up := new(UsagePeriod)

	assert.NotNil(t, up.Initialise())

	up.Name = "test-period"

	assert.NotNil(t, up.Initialise())

	up.Limit = 10

	assert.NotNil(t, up.Initialise())

	up.TimeUnit = DAY_LABEL

	assert.Nil(t, up.Initialise())
}

func TestUnitInvalidTimeUnits(t *testing.T) {

	up := new(UsagePeriod)

	up.Name = "test-period"
	up.TimeUnit = "DAYZ"

	assert.NotNil(t, up.Initialise())

	up.TimeUnit = DAY_LABEL
	up.Granularity = "MINZ"

	assert.NotNil(t, up.Initialise())
}

func TestDayClockAlignedHourGranularity(t *testing.T) {
	up := new(UsagePeriod)

	up.Name = "test-period"
	up.TimeUnit = DAY_LABEL
	up.Granularity = HOUR_LABEL
	up.ClockAlign = true
	up.Limit = 5

	mts := new(mockTimeSource)
	up.timeFunc = mts.Time
	mts.times = []string{
		"2006-01-02T15:04:05.999999999Z",
		"2006-01-02T16:05:05.999999999Z",
		"2006-01-02T17:06:05.999999999Z",
		"2006-01-02T18:07:05.999999999Z",
		"2006-01-02T19:08:05.999999999Z",
		"2006-01-02T20:09:05.999999999Z",
		"2006-01-03T15:09:05.999999999Z",
	}
	_ = up.Initialise()

	for i := 0; i < 5; i++ {
		okay, _ := up.Allowed()
		assert.True(t, okay)
		assert.Equal(t, i+1, up.TotalSlices())
	}

	//Sixth request on same day
	okay, _ := up.Allowed()
	assert.False(t, okay)

	//First request on following day
	okay, _ = up.Allowed()
	assert.True(t, okay)
	assert.Equal(t, 1, up.TotalSlices())

}

func TestDayAlignedPeriod(t *testing.T) {

	up := new(UsagePeriod)

	up.Name = "test-period"
	up.TimeUnit = DAY_LABEL
	up.Granularity = DAY_LABEL
	up.ClockAlign = true
	up.Limit = 5

	mts := new(mockTimeSource)
	up.timeFunc = mts.Time
	mts.times = []string{
		"2006-01-02T15:04:05.999999999Z",
		"2006-01-02T15:05:05.999999999Z",
		"2006-01-02T15:06:05.999999999Z",
		"2006-01-02T15:07:05.999999999Z",
		"2006-01-02T15:08:05.999999999Z",
		"2006-01-02T15:09:05.999999999Z",
		"2006-01-03T15:09:05.999999999Z",
	}
	//2006-01-02T15:04:05.999999999Z07:00

	_ = up.Initialise()

	okay, _ := up.Allowed()

	assert.True(t, okay)

	assert.Equal(t, uint64(1), up.count)
	assert.NotNil(t, up.earliest)
	assert.NotNil(t, up.latest)

	assert.Equal(t, time.Hour*24, up.duration)

	assert.Equal(t, "00:00:00.000", formatTime(up.earliest.Start))
	assert.Equal(t, "23:59:59.999", formatTime(up.earliest.End))

	for i := 0; i < 4; i++ {
		okay, _ = up.Allowed()
		assert.True(t, okay)
	}

	//Sixth request in period
	okay, _ = up.Allowed()
	assert.False(t, okay)

	assert.Equal(t, 1, up.TotalSlices())

	//First request in new period
	okay, _ = up.Allowed()
	assert.True(t, okay)

	assert.Equal(t, 1, up.TotalSlices())

}

func formatTime(t time.Time) string {
	return t.Format("15:04:05.000")
}

type mockTimeSource struct {
	times []string
}

func (mts *mockTimeSource) Time() time.Time {

	t := mts.times[0]

	if len(mts.times) > 1 {
		mts.times = mts.times[1:]
	} else {
		mts.times = nil
	}

	ft, err := time.Parse(time.RFC3339Nano, t)

	if err != nil {
		fmt.Println(err)
	}

	return ft
}
