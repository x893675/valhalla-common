package mfa

import (
	"context"
	"fmt"

	"github.com/x893675/valhalla-common/authentication/user"
	"github.com/x893675/valhalla-common/cache"
	"github.com/x893675/valhalla-common/errdetails"
	"github.com/x893675/valhalla-common/logger"
)

type Options struct {
	Providers []ProviderOption `json:"providers" yaml:"providers" toml:"providers"`
}

type ProviderOption struct {
	Type    string                 `json:"type" yaml:"type" toml:"type"`
	Options map[string]interface{} `json:"options" yaml:"options" toml:"options"`
}

type Bind interface {
	// SendBindDeviceRequest binds a device to user
	// otp: 生成绑定设备的二维码
	// sms: 发送短信确认的短信验证码
	// email: 发送邮箱确认的邮件，邮件中包含链接
	SendBindDeviceRequest(ctx context.Context, user user.Info) (string, error)
	// VerifyBindDevice verifies the bind request
	// otp: 验证一次绑定设备的6位数字
	// sms: 验证短信验证码
	// email: 验证邮箱确认的邮件
	VerifyBindDevice(ctx context.Context, user user.Info, code string) (bool, user.Info, error)
}

type TokenManager interface {
	// IssueTo issues a token to user
	// otp: 不需要
	// sms: 发送邮件，生成短信验证码
	// email: 发送邮件，生成邮件确认链接
	IssueTo(ctx context.Context, user user.Info) (string, error)
	// AuthenticationToken verifies the token
	// otp: 验证一次绑定设备的6位数字
	// sms: 验证短信验证码
	// email: 验证邮箱确认的邮件
	AuthenticationToken(ctx context.Context, user user.Info, token string, secret string) (user.Info, error)
}

type Authenticator interface {
	Bind
	TokenManager
}

type AuthenticatorFactory interface {
	Type() string
	Create(cache cache.Interface, options map[string]interface{}) (Authenticator, error)
}

var (
	mfaAuthenticatorFactories = make(map[string]AuthenticatorFactory)
	mfaAuthenticators         = make(map[string]Authenticator)
)

func RegisterAuthenticatorFactory(factory AuthenticatorFactory) {
	kind := factory.Type()
	if _, ok := mfaAuthenticatorFactories[kind]; ok {
		panic(fmt.Errorf("already registered type: %s", kind))
	}
	mfaAuthenticatorFactories[kind] = factory
}

func SetupWithOptions(p cache.Interface, opts *Options) error {
	if opts == nil || len(opts.Providers) == 0 {
		return nil
	}
	for _, o := range opts.Providers {
		if mfaAuthenticators[o.Type] != nil {
			return fmt.Errorf("duplicate mfa authenticator type found: %s", o.Type)
		}
		if mfaAuthenticatorFactories[o.Type] == nil {
			return fmt.Errorf("mfa authenticator %s is not supported", o.Type)
		}
		if factory, ok := mfaAuthenticatorFactories[o.Type]; ok {
			if authenticator, err := factory.Create(p, o.Options); err != nil {
				logger.Errorf("failed to create mfa authenticator %s: %s", o.Type, err)
			} else {
				mfaAuthenticators[o.Type] = authenticator
				logger.Debugf("create mfa authenticator %s successfully", o.Type)
			}
		}
	}
	return nil
}

func SendBindDeviceRequest(ctx context.Context, user user.Info, mfaType string) (string, error) {
	if len(mfaAuthenticators) == 0 || mfaAuthenticators[mfaType] == nil {
		return "", errdetails.NotImplementedError("mfa authenticator %s is not supported", mfaType)
	}
	return mfaAuthenticators[mfaType].SendBindDeviceRequest(ctx, user)
}

func VerifyBindDevice(ctx context.Context, user user.Info, code string, mfaType string) (bool, user.Info, error) {
	if len(mfaAuthenticators) == 0 || mfaAuthenticators[mfaType] == nil {
		return false, user, errdetails.NotImplementedError("mfa authenticator %s is not supported", mfaType)
	}
	return mfaAuthenticators[mfaType].VerifyBindDevice(ctx, user, code)
}

func IssueTo(ctx context.Context, user user.Info, mfaType string) (string, error) {
	if len(mfaAuthenticators) == 0 || mfaAuthenticators[mfaType] == nil {
		return "", errdetails.NotImplementedError("mfa authenticator %s is not supported", mfaType)
	}
	return mfaAuthenticators[mfaType].IssueTo(ctx, user)
}

func AuthenticationToken(ctx context.Context, user user.Info, token string, mfaType string, secret string) (user.Info, error) {
	if len(mfaAuthenticators) == 0 || mfaAuthenticators[mfaType] == nil {
		return nil, errdetails.NotImplementedError("mfa authenticator %s is not supported", mfaType)
	}
	return mfaAuthenticators[mfaType].AuthenticationToken(ctx, user, token, secret)
}
