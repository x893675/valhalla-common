package token

import (
	"context"
	"errors"

	"github.com/x893675/valhalla-common/authentication/authenticator"
	"github.com/x893675/valhalla-common/authentication/user"
)

var _ authenticator.Token = (*SystemAccountTokenAuthenticator)(nil)

// ErrInvalidSystemAccountToken indicates the token is missing or not a valid system service account credential.
// Resolvers should return this for "not found" so AuthenticateToken can return (nil, false, nil) and union auth can fall through.
var ErrInvalidSystemAccountToken = errors.New("invalid system account token")

// SystemAccountResolver resolves a raw bearer token to user.Info for system service accounts (typically DB-backed).
type SystemAccountResolver interface {
	Resolve(ctx context.Context, token string) (user.Info, error)
}

// SystemAccountTokenAuthenticator validates opaque tokens issued to system service accounts via SystemAccountResolver.
type SystemAccountTokenAuthenticator struct {
	resolve SystemAccountResolver
}

// NewSystemAccountTokenAuthenticator constructs an authenticator.Token for system service account bearer tokens.
func NewSystemAccountTokenAuthenticator(r SystemAccountResolver) *SystemAccountTokenAuthenticator {
	return &SystemAccountTokenAuthenticator{resolve: r}
}

// Verify resolves the token to a user identity.
func (a *SystemAccountTokenAuthenticator) Verify(ctx context.Context, token string) (user.Info, error) {
	if a.resolve == nil {
		return nil, errors.New("system account resolver is nil")
	}
	if token == "" {
		return nil, ErrInvalidSystemAccountToken
	}
	u, err := a.resolve.Resolve(ctx, token)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrInvalidSystemAccountToken
	}
	return u, nil
}

// AuthenticateToken implements authenticator.Token.
func (a *SystemAccountTokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	u, err := a.Verify(ctx, token)
	if err != nil {
		if errors.Is(err, ErrInvalidSystemAccountToken) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &authenticator.Response{User: u}, true, nil
}
