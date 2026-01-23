package policy

import (
	"net"
	"net/http"
)

var _ ConditionParser = (*SourceIP)(nil)

/*
SourceIP

	{
		"acs:SourceIp": ["10.0.0.1", "192.168.1.1/16"]
	}
*/
type SourceIP struct{}

const (
	XForwardedFor = "X-Forwarded-For"
	XRealIP       = "X-Real-IP"
	XClientIP     = "x-client-ip"
)

func (c *SourceIP) ParseCondition(req *http.Request) any {
	remoteAddr := req.RemoteAddr
	if ip := req.Header.Get(XClientIP); ip != "" {
		remoteAddr = ip
	} else if ip := req.Header.Get(XRealIP); ip != "" {
		remoteAddr = ip
	} else if ip = req.Header.Get(XForwardedFor); ip != "" {
		remoteAddr = ip
	} else {
		remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
	}

	if remoteAddr == "::1" {
		remoteAddr = "127.0.0.1"
	}

	return remoteAddr
}
