package accesstoken

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/x893675/valhalla-common/authentication/authenticator"
)

var _ authenticator.Request = (*Authenticator)(nil)

var ErrInvalidToken = errors.New("invalid access token")

// Authenticator implements authenticator.Request
// Authorization: Token <token>
type Authenticator struct {
	auth authenticator.Token
}

func (a *Authenticator) AuthenticateRequest(req *http.Request) (*authenticator.Response, bool, error) {
	auth := strings.TrimSpace(req.Header.Get("Authorization"))
	if auth == "" {
		return nil, false, fmt.Errorf("[access_token] authorization in header is empty")
	}
	parts := strings.Split(auth, " ")
	if len(parts) < 2 || strings.ToLower(parts[0]) != "token" {
		return nil, false, fmt.Errorf("[access_token] token[%s] format error", auth)
	}

	token := parts[1]

	// Empty access tokens aren't valid
	if len(token) == 0 {
		return nil, false, fmt.Errorf("[access token]  token[%s] is empty", auth)
	}

	resp, ok, err := a.auth.AuthenticateToken(req.Context(), token)

	// If the token authenticator didn't error, provide a default error
	if !ok && err == nil {
		err = ErrInvalidToken
	}

	return resp, ok, err
}

func New(auth authenticator.Token) authenticator.Request {
	return &Authenticator{auth: auth}
}
