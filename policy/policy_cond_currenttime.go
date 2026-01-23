package policy

import (
	"net/http"
	"time"
)

var _ ConditionParser = (*CurrentTime)(nil)

/*
CurrentTime

	{
		"acs:CurrentTime": "2019-01-01T00:00:00Z08:00"
	}
*/
type CurrentTime struct{}

func (c *CurrentTime) ParseCondition(_ *http.Request) any {
	return time.Now().UTC().Format(time.RFC3339)
}
