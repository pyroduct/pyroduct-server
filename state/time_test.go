package state

import (
	"github.com/stretchr/testify/assert"
	"testing"
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

	assert.Nil(t, parseAndValidateUsagePeriod(up))

	up.Granularity = HOUR_LABEL
	assert.Nil(t, parseAndValidateUsagePeriod(up))

	assert.Equal(t, TimeUnit(DAY), up.timeUnitEnum)
	assert.Equal(t, TimeUnit(HOUR), up.granularityEnum)
	assert.Equal(t, 1, up.UnitMultiple)

}

func TestNameValidation(t *testing.T) {

	up := new(UsagePeriod)

	up.Name = "test-period"
	up.TimeUnit = DAY_LABEL

	assert.Nil(t, parseAndValidateUsagePeriod(up))

	up.Name = "test period"
	assert.NotNil(t, parseAndValidateUsagePeriod(up))

	up.Name = "test\tperiod"
	assert.NotNil(t, parseAndValidateUsagePeriod(up))

	up.Name = "test\nperiod"
	assert.NotNil(t, parseAndValidateUsagePeriod(up))
}

func TestUnitGranularityRelationship(t *testing.T) {

	up := new(UsagePeriod)

	up.Name = "test-period"
	up.TimeUnit = DAY_LABEL
	up.Granularity = DAY_LABEL

	assert.Nil(t, parseAndValidateUsagePeriod(up))

	up.Granularity = HOUR_LABEL

	assert.Nil(t, parseAndValidateUsagePeriod(up))

	up.TimeUnit = SECOND_LABEL
	up.Granularity = MINUTE_LABEL

	assert.NotNil(t, parseAndValidateUsagePeriod(up))

}

func TestUnitInvalidUnitMultiple(t *testing.T) {

	up := new(UsagePeriod)

	up.Name = "test-period"
	up.TimeUnit = DAY_LABEL
	up.UnitMultiple = -1

	assert.NotNil(t, parseAndValidateUsagePeriod(up))
}

func TestMissingMandatory(t *testing.T) {

	up := new(UsagePeriod)

	assert.NotNil(t, parseAndValidateUsagePeriod(up))

	up.Name = "test-period"

	assert.NotNil(t, parseAndValidateUsagePeriod(up))

	up.TimeUnit = DAY_LABEL

	assert.Nil(t, parseAndValidateUsagePeriod(up))
}

func TestUnitInvalidTimeUnits(t *testing.T) {

	up := new(UsagePeriod)

	up.Name = "test-period"
	up.TimeUnit = "DAYZ"

	assert.NotNil(t, parseAndValidateUsagePeriod(up))

	up.TimeUnit = DAY_LABEL
	up.Granularity = "MINZ"

	assert.NotNil(t, parseAndValidateUsagePeriod(up))
}
