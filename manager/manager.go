package manager

import (
	"fmt"
	"github.com/graniticio/granitic/v2/logging"
	"github.com/pyroduct/pyroduct-server/state"
)

const (
	timePeriod  = "timePeriod"
	alignMode   = "CLOCK_ALIGNED"
	rollingMode = "ROLLING"
)

type PyroductManager struct {
	Config      map[string]interface{}
	Log         logging.Logger
	managedApis map[string]*ManagedApi
}

func (pm *PyroductManager) StartComponent() error {

	pm.Log.LogDebugf("Validating configuration file")

	pm.managedApis = make(map[string]*ManagedApi)

	if errors := pm.buildFromConfig(); len(errors) > 0 {

		for _, err := range errors {
			pm.Log.LogErrorf(err.Error())
		}

		return fmt.Errorf("unable to create API management rules from configuration")

	}

	if len(pm.managedApis) == 0 {
		pm.Log.LogWarnf("Found zero rule sets in configuration")
	} else {
		pm.Log.LogInfof("Found %d rule set(s) in configuration", len(pm.managedApis))
	}

	return nil

}

func (pm *PyroductManager) buildFromConfig() []error {

	var errors []error

	l := pm.Log

	for name, config := range pm.Config {

		ma := new(ManagedApi)

		l.LogDebugf("Parsing config for %s", name)

		upc := config.(map[string]interface{})[timePeriod]

		if upc == nil {
			//No usage period config
			e := fmt.Errorf("no %s config block defined for %s", timePeriod, name)
			errors = append(errors, e)
			continue
		}

		if us, ne := pm.usagePeriodFromConfig(name, upc.(map[string]interface{})); len(ne) > 0 {
			errors = append(errors, ne...)
		} else {
			ma.UsagePeriod = us
			ma.Name = name

			pm.managedApis[name] = ma
		}

	}

	return errors

}

func (pm *PyroductManager) usagePeriodFromConfig(name string, config map[string]interface{}) (*state.UsagePeriod, []error) {

	var errors []error
	us := new(state.UsagePeriod)

	errors = append(errors, pm.extractAndValidateMode(name, config, errors, us)...)
	errors = append(errors, pm.extractAndValidateUnit(name, config, errors, us)...)
	errors = append(errors, pm.extractAndValidateLimit(name, config, errors, us)...)
	errors = append(errors, pm.extractAndValidateQuantity(name, config, errors, us)...)

	return us, errors
}

func (pm *PyroductManager) extractAndValidateMode(name string, config map[string]interface{}, errors []error, us *state.UsagePeriod) []error {
	mode, okay := config["mode"].(string)

	if !okay {
		errors = append(errors, fmt.Errorf("%s.%s.mode not defined (or value is not a string)", name, timePeriod))
	} else if mode != rollingMode && mode != alignMode {
		errors = append(errors, fmt.Errorf("%s.%s.mode can only be %s or %s (is %s)", name, timePeriod, rollingMode, alignMode, mode))

	} else {
		us.ClockAlign = mode == alignMode
	}
	return errors
}

func (pm *PyroductManager) extractAndValidateUnit(name string, config map[string]interface{}, errors []error, us *state.UsagePeriod) []error {
	unit, okay := config["unit"].(string)

	if !okay {
		errors = append(errors, fmt.Errorf("%s.%s.unit not defined (or value is not a string)", name, timePeriod))
	} else if unit != state.DAY_LABEL && unit != state.HOUR_LABEL &&
		unit != state.MINUTE_LABEL && unit != state.SECOND_LABEL {

		errors = append(errors, fmt.Errorf("%s.%s.unit can only be %s, %s, %s, or %s (is %s)", name, timePeriod,
			state.DAY_LABEL, state.HOUR_LABEL, state.MINUTE_LABEL, state.SECOND_LABEL, unit))

	} else {
		us.TimeUnit = unit
	}
	return errors
}

func (pm *PyroductManager) extractAndValidateLimit(name string, config map[string]interface{}, errors []error, us *state.UsagePeriod) []error {
	limit, okay := config["requestLimit"].(int)

	if !okay {
		errors = append(errors, fmt.Errorf("%s.%s.requestLimit not defined (or value is not a positive integer)", name, timePeriod))
	} else if limit <= 0 {
		errors = append(errors, fmt.Errorf("%s.%s.requestLimit must be a positive integer", name, timePeriod))
	} else {
		us.Limit = uint64(limit)
	}
	return errors
}

func (pm *PyroductManager) extractAndValidateQuantity(name string, config map[string]interface{}, errors []error, us *state.UsagePeriod) []error {

	q, ok := config["quantity"]

	if !ok {
		//Unset - default to 1
		us.UnitMultiple = 1
		return nil
	}

	qi, okay := q.(int)

	if !okay || qi <= 0 {
		errors = append(errors, fmt.Errorf("if set, %s.%s.quantity must be a positive integer", name, timePeriod))
	}

	return errors
}

type ManagedApi struct {
	Name string
	*state.UsagePeriod
}
