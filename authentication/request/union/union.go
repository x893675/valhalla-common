package union

import (
	"errors"
	"net/http"

	"github.com/x893675/valhalla-common/authentication/authenticator"
)

var _ authenticator.Request = (*unionAuthRequestHandler)(nil)

type unionAuthRequestHandler struct {
	Handlers    []authenticator.Request
	FailOnError bool
}

func (u *unionAuthRequestHandler) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	var errlist []error
	for _, currAuthRequestHandler := range u.Handlers {
		resp, ok, err := currAuthRequestHandler.AuthenticateRequest(req)
		if err != nil {
			if u.FailOnError {
				return resp, ok, err
			}
			errlist = append(errlist, err)
			continue
		}

		if ok {
			return resp, ok, err
		}
	}

	return nil, false, errors.Join(errlist...)
}

// New returns a request authenticator that validates credentials using a chain of authenticator.Request objects.
// The entire chain is tried until one succeeds. If all fail, an aggregate error is returned.
func New(authRequestHandlers ...authenticator.Request) authenticator.Request {
	if len(authRequestHandlers) == 1 {
		return authRequestHandlers[0]
	}
	return &unionAuthRequestHandler{Handlers: authRequestHandlers, FailOnError: false}
}

// NewFailOnError returns a request authenticator that validates credentials using a chain of authenticator.Request objects.
// The first error short-circuits the chain.
func NewFailOnError(authRequestHandlers ...authenticator.Request) authenticator.Request {
	if len(authRequestHandlers) == 1 {
		return authRequestHandlers[0]
	}
	return &unionAuthRequestHandler{Handlers: authRequestHandlers, FailOnError: true}
}
