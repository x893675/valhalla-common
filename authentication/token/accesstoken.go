package token

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/x893675/valhalla-common/authentication/authenticator"
	"github.com/x893675/valhalla-common/authentication/user"
	"github.com/x893675/valhalla-common/cache"
	"github.com/x893675/valhalla-common/constant"
	"github.com/x893675/valhalla-common/errdetails"
	"github.com/x893675/valhalla-common/logger"
	"github.com/x893675/valhalla-common/utils/crypto"
)

var _ authenticator.Token = (*AESTokenAuthenticator)(nil)
var _ TokenManager = (*AESTokenAuthenticator)(nil)

// ErrInvalidSystemAccountToken is returned by SystemAccountResolver when the token is unknown or invalid.
var ErrInvalidSystemAccountToken = errors.New("invalid system account token")

// SystemAccountResolver checks a bearer token against persistent storage for system service accounts (e.g. DB).
type SystemAccountResolver interface {
	Resolve(ctx context.Context, token string) (user.Info, error)
}

// Claims is the payload embedded in AES access tokens.
type Claims struct {
	UID       string `json:"uid"`
	ExpiresAt int64  `json:"exp,omitempty"`
	Issuer    string `json:"iss,omitempty"`
	// Ut is user.UserType as string (e.g. "account", "service_account"). Empty means legacy tokens that only use cache lookup.
	Ut string `json:"ut,omitempty"`
}

type AESTokenAuthenticator struct {
	secret      []byte
	cache       cache.Interface
	now         func() time.Time
	ssaResolver SystemAccountResolver
}

func (a *AESTokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	u, err := a.verify(ctx, token)
	if err != nil {
		return nil, false, err
	}
	return &authenticator.Response{
		User: u,
	}, true, nil
}

// Verify validates a token without request context (uses [context.Background] for storage resolution). Prefer AuthenticateToken when context is available.
func (a *AESTokenAuthenticator) Verify(token string) (user.Info, error) {
	return a.verify(context.Background(), token)
}

func (a *AESTokenAuthenticator) verify(ctx context.Context, wireToken string) (user.Info, error) {
	logger.Debugf("Verifying access token: %s", wireToken)
	if wireToken == "" {
		return nil, fmt.Errorf("token is empty")
	}
	claim, err := a.parseClaims(wireToken)
	if err != nil {
		return a.verifyOpaqueServiceAccount(ctx, wireToken, err)
	}
	now := a.now().UTC().Unix()
	if now > claim.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}
	if claim.Ut == string(user.UserTypeServiceAccount) {
		return a.verifyServiceAccount(ctx, wireToken)
	}
	u := user.DefaultInfo{}
	if err := a.cache.Get(context.TODO(), fmt.Sprintf(constant.TokenCacheKeyFormat, claim.UID, wireToken), &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (a *AESTokenAuthenticator) parseClaims(wireToken string) (*Claims, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(wireToken)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) == 0 {
		return nil, fmt.Errorf("token is invalid")
	}
	plaintext, err := crypto.AESCBCDecrypt(ciphertext, a.secret)
	if err != nil {
		return nil, err
	}
	claim := Claims{}
	if err := json.Unmarshal(plaintext, &claim); err != nil {
		return nil, err
	}
	return &claim, nil
}

func (a *AESTokenAuthenticator) verifyServiceAccount(ctx context.Context, wireToken string) (user.Info, error) {
	if a.ssaResolver == nil {
		return nil, fmt.Errorf("service account resolver is not configured")
	}
	u, err := a.ssaResolver.Resolve(ctx, wireToken)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, ErrInvalidSystemAccountToken
	}
	return u, nil
}

// verifyOpaqueServiceAccount handles legacy non-AES tokens stored as opaque secrets in storage.
func (a *AESTokenAuthenticator) verifyOpaqueServiceAccount(ctx context.Context, token string, aesErr error) (user.Info, error) {
	if a.ssaResolver == nil {
		return nil, aesErr
	}
	u, err := a.ssaResolver.Resolve(ctx, token)
	if err != nil {
		if errors.Is(err, ErrInvalidSystemAccountToken) {
			return nil, aesErr
		}
		return nil, err
	}
	if u == nil {
		return nil, aesErr
	}
	return u, nil
}

func (a *AESTokenAuthenticator) IssueTo(ctx context.Context, u user.Info, expire time.Duration) (string, error) {
	expirein := a.now().UTC().Add(expire).Unix()
	ut := ""
	if u != nil {
		ut = string(u.UserType())
	}
	claim := Claims{
		UID:       u.GetID(),
		ExpiresAt: expirein,
		Issuer:    "valhalla",
		Ut:        ut,
	}
	claimBytes, err := json.Marshal(claim)
	if err != nil {
		return "", err
	}
	ciphertext, err := crypto.AESCBCEncrypt(claimBytes, a.secret)
	if err != nil {
		return "", err
	}
	t := base64.URLEncoding.EncodeToString(ciphertext)
	if err := a.cache.Set(ctx, fmt.Sprintf(constant.TokenCacheKeyFormat, u.GetID(), t), u, expire); err != nil {
		return "", errdetails.CacheOperationFailed("cache token operation failed: %v", err)
	}
	return t, nil
}

func (a *AESTokenAuthenticator) RevokeAllUserTokens(ctx context.Context, uid string) error {
	return a.cache.RemoveWithPattern(ctx, fmt.Sprintf(constant.TokenCacheKeyFormat, uid, "*"))
}

// NewAESTokenAuthenticator builds the unified access token authenticator. ssa may be nil if system service accounts are not used.
func NewAESTokenAuthenticator(secret []byte, cache cache.Interface, fn func() time.Time, ssa SystemAccountResolver) *AESTokenAuthenticator {
	return &AESTokenAuthenticator{
		cache:       cache,
		secret:      secret,
		now:         fn,
		ssaResolver: ssa,
	}
}
