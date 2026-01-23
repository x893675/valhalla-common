package policy

import (
	"encoding/json"
)

func ConditionMather(arguments ...interface{}) (interface{}, error) {
	condsContextString := arguments[0].(string)
	conditionString := arguments[1].(string)
	var conds Condition
	err := json.Unmarshal([]byte(conditionString), &conds)
	if err != nil {
		return false, err
	}
	var condsContext ConditionContext
	err = json.Unmarshal([]byte(condsContextString), &condsContext)
	if err != nil {
		return false, err
	}

	for k, cond := range conds {
		fn, ok := conditionOperatorFuncMap[k]
		if !ok {
			return false, nil
		}
		for condKey, v1 := range cond {
			if _, ok := condsContext[condKey]; !ok {
				return false, nil
			}
			if !fn(condsContext[condKey], v1) {
				return false, nil
			}
		}
	}
	return true, nil
}
