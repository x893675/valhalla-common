package policy

import "net/http"

var _ ConditionParser = (*Service)(nil)

/*
Service

	{
		"acs:Service": ["ecs.aliyuncs.com"]
	}
*/
type Service struct{}

const (
	XServiceName = "X-Service-Name"
)

func (c *Service) ParseCondition(req *http.Request) any {
	return req.Header.Get(XServiceName)
}
