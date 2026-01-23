package token

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/x893675/valhalla-common/authentication/authenticator"
	"github.com/x893675/valhalla-common/authentication/user"
	"github.com/x893675/valhalla-common/cache"
	"github.com/x893675/valhalla-common/constant"
	"github.com/x893675/valhalla-common/errdetails"
	"github.com/x893675/valhalla-common/utils/crypto"
)

var _ authenticator.Token = (*AESTokenAuthenticator)(nil)
var _ TokenManager = (*AESTokenAuthenticator)(nil)

type Claims struct {
	UID       uint64 `json:"uid"`
	ExpiresAt int64  `json:"exp,omitempty"`
	Issuer    string `json:"iss,omitempty"`
}

type AESTokenAuthenticator struct {
	secret []byte
	cache  cache.Interface
	now    func() time.Time
}

func (a *AESTokenAuthenticator) AuthenticateToken(ctx context.Context, token string) (*authenticator.Response, bool, error) {
	u, err := a.Verify(token)
	if err != nil {
		return nil, false, err
	}
	return &authenticator.Response{
		User: u,
	}, true, nil
}

func (a *AESTokenAuthenticator) Verify(token string) (user.Info, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return nil, err
	}
	plaintext, err := crypto.AESCBCDecrypt(ciphertext, a.secret)
	if err != nil {
		return nil, err
	}
	claim := Claims{}
	if err := json.Unmarshal(plaintext, &claim); err != nil {
		return nil, err
	}
	now := a.now().UTC().Unix()
	if now > claim.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}
	u := user.DefaultInfo{}
	if err := a.cache.Get(context.TODO(), fmt.Sprintf(constant.TokenCacheKeyFormat, claim.UID, token), &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (a *AESTokenAuthenticator) IssueTo(ctx context.Context, user user.Info, expire time.Duration) (string, error) {
	expirein := a.now().UTC().Add(expire).Unix()
	claim := Claims{
		UID:       user.GetID(),
		ExpiresAt: expirein,
		Issuer:    "valhalla",
	}
	claimBytes, _ := json.Marshal(claim)
	ciphertext, err := crypto.AESCBCEncrypt(claimBytes, a.secret)
	if err != nil {
		return "", err
	}
	t := base64.URLEncoding.EncodeToString(ciphertext)
	if err := a.cache.Set(ctx, fmt.Sprintf(constant.TokenCacheKeyFormat, user.GetID(), t), user, expire); err != nil {
		return "", errdetails.CacheOperationFailed("cache token operation failed: %v", err)
	}
	return t, nil
}

func (a *AESTokenAuthenticator) RevokeAllUserTokens(ctx context.Context, uid uint64) error {
	return a.cache.RemoveWithPattern(ctx, fmt.Sprintf(constant.TokenCacheKeyFormat, uid, "*"))
}

func NewAESTokenAuthenticator(secret []byte, cache cache.Interface, fn func() time.Time) *AESTokenAuthenticator {
	return &AESTokenAuthenticator{
		cache:  cache,
		secret: secret,
		now:    fn,
	}
}
