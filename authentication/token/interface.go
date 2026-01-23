package token

import (
	"context"
	"fmt"
	"time"

	"github.com/x893675/valhalla-common/authentication/authenticator"
	"github.com/x893675/valhalla-common/authentication/user"
	"github.com/x893675/valhalla-common/cache"
	"github.com/x893675/valhalla-common/logger"
)

type TokenManager interface {
	authenticator.Token
	// IssueTo issues a token a User, return error if issuing process failed
	IssueTo(ctx context.Context, user user.Info, expire time.Duration) (string, error)
	// RevokeAllUserTokens revoke all user tokens
	RevokeAllUserTokens(ctx context.Context, uid uint64) error
}

type Options struct {
	Type   string `json:"type" yaml:"type"`
	Secret string `json:"secret" yaml:"secret"`
}

func DefaultOptions() *Options {
	return &Options{
		Type:   "aes",
		Secret: "12345678abcdefgh12345678abcdefgh", //aes-256
	}
}

func NewTokenManager(cache cache.Interface, opts *Options) (TokenManager, error) {
	if opts == nil {
		logger.Debug("token manager options is nil, use default options")
		opts = DefaultOptions()
	}
	switch opts.Type {
	case "aes":
		return NewAESTokenAuthenticator([]byte(opts.Secret), cache, time.Now), nil
	default:
		return nil, fmt.Errorf("unknown token type: %s", opts.Type)
	}
}
