package mfa

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/gomail.v2"

	"github.com/x893675/valhalla-common/authentication/user"
	"github.com/x893675/valhalla-common/cache"
	"github.com/x893675/valhalla-common/constant"
	"github.com/x893675/valhalla-common/errdetails"
	"github.com/x893675/valhalla-common/logger"
	"github.com/x893675/valhalla-common/utils/random"
)

const verifyEmailTemplate = `
<html>
<body>
<h3>%s , 您好</h3>
<p>请使用您的验证码进行验证：</p>
<a style="background-color: green; color: white; padding: 10px 20px; text-decoration: none;">%s</a>
</body>
</html>
`

const sendVerificationCodeTemplate = `
<html>
<body>
<h3>%s , 您好</h3>
<p>您的验证码是：<b>%s</b></p>
</body>
</html>
`

func init() {
	RegisterAuthenticatorFactory(&SMTPProviderFactory{})
}

type SMTPProviderFactory struct{}

func (s *SMTPProviderFactory) Type() string {
	return constant.MFAProviderEmail
}

func (s *SMTPProviderFactory) Create(cache cache.Interface, options map[string]interface{}) (Authenticator, error) {
	var smtp SMTPProvider

	if err := mapstructure.Decode(options, &smtp); err != nil {
		return nil, err
	}
	smtp.cache = cache
	if smtp.Port == 0 {
		smtp.Port = 25
	}
	if smtp.From == "" {
		return nil, fmt.Errorf("from is required")
	}
	if smtp.SmartHost == "" {
		return nil, fmt.Errorf("smart_host is required")
	}
	if smtp.CacheExpire == "" {
		smtp.expire = constant.MFATokenCacheDuration
	} else {
		d, err := time.ParseDuration(smtp.CacheExpire)
		if err != nil {
			logger.Errorf("failed to parse cache expire duration: %s", err)
			return nil, err
		}
		smtp.expire = d
	}
	smtp.smtp = gomail.NewDialer(smtp.SmartHost, smtp.Port, smtp.Username, smtp.Password)
	return &smtp, nil
}

type SMTPProvider struct {
	Username  string `json:"username" yaml:"username"`
	Password  string `json:"password" yaml:"password"`
	SmartHost string `json:"smartHost" yaml:"smartHost"`
	Port      int    `json:"port" yaml:"port"`
	Insecure  bool   `json:"insecure" yaml:"insecure"`
	From      string `json:"from" yaml:"from"`
	//RedirectURL string `json:"redirectURL" yaml:"redirectURL"`
	CacheExpire string `json:"cacheExpire" yaml:"cacheExpire"`
	smtp        *gomail.Dialer
	expire      time.Duration
	cache       cache.Interface
}

// VerifyBindDevice verifies the bind device request.
// 跟 totp 不同， totp 是在已登录状态下，生成密钥，让用户扫码，再验证一次，全程是在登录状态下， API 过来之后知道用户是谁
// 邮件验证是向用户邮箱发送验证链接，用户点击链接之后，直接更改状态，链接跳转不携带用户信息
func (s *SMTPProvider) VerifyBindDevice(ctx context.Context, iuser user.Info, code string) (bool, user.Info, error) {
	var cacheUser user.DefaultInfo
	if err := s.cache.Get(ctx, fmt.Sprintf(constant.EmailBindCacheKeyFormat, iuser.GetID(), code), &cacheUser); err != nil {
		if errors.Is(err, cache.ErrNotExists) {
			return false, nil, nil
		}
		logger.Errorf("failed to get user from cache: %s", err)
		return false, nil, err
	}
	go func() {
		if err := s.cache.Remove(context.TODO(), fmt.Sprintf(constant.EmailBindCacheKeyFormat, iuser.GetID(), code)); err != nil {
			logger.Warnf("failed to remove email bind code from cache: %s", err)
		}
	}()
	return true, &cacheUser, nil
}

func (s *SMTPProvider) IssueTo(ctx context.Context, user user.Info) (string, error) {
	code := random.RandDigitString(6)
	msg := gomail.NewMessage()
	msg.SetHeader("From", s.From)
	msg.SetHeader("To", user.GetEmail())
	msg.SetHeader("Subject", "您的验证码")
	msg.SetBody("text/html", fmt.Sprintf(sendVerificationCodeTemplate, user.GetName(), code))
	if err := s.cache.Set(ctx, fmt.Sprintf(constant.EmailVerifyCacheKeyFormat, user.GetID(), code), user, s.expire); err != nil {
		logger.Errorf("failed to cache email verification code: %s", err)
		return "", errdetails.CacheOperationFailed("cache email verification code")
	}
	go func() {
		if err := s.smtp.DialAndSend(msg); err != nil {
			logger.Errorf("failed to send email: %s", err)
		}
	}()

	return code, nil
}

func (s *SMTPProvider) AuthenticationToken(ctx context.Context, iuser user.Info, token string, secret string) (user.Info, error) {
	var cacheUser user.DefaultInfo
	if err := s.cache.Get(ctx, fmt.Sprintf(constant.EmailVerifyCacheKeyFormat, iuser.GetID(), token), &cacheUser); err != nil {
		if errors.Is(err, cache.ErrNotExists) {
			return nil, errdetails.Forbidden("invalid email verification code")
		}
		logger.Errorf("failed to get user from cache: %s", err)
		return nil, err
	}
	go func() {
		if err := s.cache.Remove(context.TODO(), fmt.Sprintf(constant.EmailVerifyCacheKeyFormat, iuser.GetID(), token)); err != nil {
			logger.Warnf("failed to remove email verification code from cache: %s", err)
		}
	}()
	return &cacheUser, nil
}

func (s *SMTPProvider) SendBindDeviceRequest(ctx context.Context, user user.Info) (string, error) {
	code := random.RandDigitString(6)

	msg := gomail.NewMessage()
	msg.SetHeader("From", s.From)
	msg.SetHeader("To", user.GetEmail())
	msg.SetHeader("Subject", "请验证您的邮箱")
	//msg.SetBody("text/html", fmt.Sprintf(verifyEmailTemplate, user.GetName(), fmt.Sprintf("%s?type=%s&code=%s", s.RedirectURL, property.MFAProviderEmail, code)))
	msg.SetBody("text/html", fmt.Sprintf(verifyEmailTemplate, user.GetName(), code))
	if err := s.cache.Set(ctx, fmt.Sprintf(constant.EmailBindCacheKeyFormat, user.GetID(), code), user, s.expire); err != nil {
		logger.Errorf("failed to cache email bind code: %s", err)
		return "", err
	}

	go func() {
		if err := s.smtp.DialAndSend(msg); err != nil {
			logger.Errorf("failed to send email: %s", err)
		}
	}()

	return code, nil
}
