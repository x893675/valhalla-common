package mfa

import (
	"context"
	"errors"
	"fmt"
	"time"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"

	"github.com/x893675/valhalla-common/authentication/user"
	"github.com/x893675/valhalla-common/cache"
	"github.com/x893675/valhalla-common/constant"
	"github.com/x893675/valhalla-common/errdetails"
	"github.com/x893675/valhalla-common/logger"
	"github.com/x893675/valhalla-common/utils/random"
)

func init() {
	RegisterAuthenticatorFactory(&SMSProviderFactory{})
}

type SMSProviderFactory struct{}

func (s *SMSProviderFactory) Type() string {
	return constant.MFAProviderSMS
}

func (s *SMSProviderFactory) Create(cache cache.Interface, options map[string]interface{}) (Authenticator, error) {
	var sms SMSProvider
	if err := mapstructure.Decode(options, &sms); err != nil {
		return nil, err
	}
	if sms.AliyunSMSConfig == nil {
		return nil, fmt.Errorf("aliyun sms config is required")
	}
	if sms.CacheExpire == "" {
		sms.expire = constant.MFATokenCacheDuration
	} else {
		d, err := time.ParseDuration(sms.CacheExpire)
		if err != nil {
			logger.Errorf("failed to parse cache expire duration: %s", err)
			return nil, err
		}
		sms.expire = d
	}
	if sms.RateLimitInterval == "" {
		sms.rateLimitInterval = 1 * time.Minute
	} else {
		d, err := time.ParseDuration(sms.RateLimitInterval)
		if err != nil {
			logger.Errorf("failed to parse rate limit interval duration: %s", err)
			return nil, err
		}
		sms.rateLimitInterval = d
	}

	cfg := &openapi.Config{}
	cfg.SetAccessKeyId(sms.AliyunSMSConfig.AccessKeyID)
	cfg.SetAccessKeySecret(sms.AliyunSMSConfig.AccessKeySecret)
	cfg.SetEndpoint(sms.AliyunSMSConfig.Endpoint)

	client, err := dysmsapi.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	sms.aliyunSMSClient = client
	sms.cache = cache
	return &sms, nil
}

type AliyunSMSConfig struct {
	CodeLength      int    `json:"codeLength,omitempty" yaml:"codeLength"`
	AccessKeyID     string `json:"accessKeyID,omitempty" yaml:"accessKeyID"`
	AccessKeySecret string `json:"accessKeySecret,omitempty" yaml:"accessKeySecret"`
	Endpoint        string `json:"endpoint,omitempty" yaml:"endpoint"`
	SignName        string `json:"signName,omitempty" yaml:"signName"`
	TemplateCode    string `json:"templateCode,omitempty" yaml:"templateCode"`
}

type SMSProvider struct {
	AliyunSMSConfig   *AliyunSMSConfig `json:"aliyunSMSConfig" yaml:"aliyunSMSConfig"`
	CacheExpire       string           `json:"cacheExpire" yaml:"cacheExpire"`
	RateLimitInterval string           `json:"rateLimitInterval" yaml:"rateLimitInterval"`
	aliyunSMSClient   *dysmsapi.Client
	expire            time.Duration
	rateLimitInterval time.Duration
	cache             cache.Interface
}

func (s *SMSProvider) SendBindDeviceRequest(ctx context.Context, user user.Info) (string, error) {
	exist, err := s.cache.Exist(ctx, fmt.Sprintf(constant.SMSBindRateLimitKeyFormat, user.GetID()))
	if err != nil {
		logger.Errorf("failed to check rate limit: %s", err)
		return "", err
	}
	if exist {
		return "", errdetails.SendSMSTooFrequently("send sms too frequently, retry after %v sec", s.rateLimitInterval.Seconds())
	}

	code := random.RandDigitString(s.AliyunSMSConfig.CodeLength)

	if err := s.cache.Set(ctx, fmt.Sprintf(constant.SMSBindCacheKeyFormat, user.GetID(), code), user, s.expire); err != nil {
		logger.Errorf("failed to cache sms bind code: %s", err)
		return "", err
	}

	go func() {
		if err := s.cache.Set(ctx, fmt.Sprintf(constant.SMSBindRateLimitKeyFormat, user.GetID()), "", s.rateLimitInterval); err != nil {
			logger.Errorf("failed to cache email bind rate limit: %s", err)
		}
	}()

	go func() {
		req := dysmsapi.SendSmsRequest{}
		req.SetSignName(s.AliyunSMSConfig.SignName)
		req.SetTemplateCode(s.AliyunSMSConfig.TemplateCode)
		req.SetPhoneNumbers(user.GetPhone())
		req.SetTemplateParam(fmt.Sprintf("{\"code\":\"%s\"}", code))
		_, err := s.aliyunSMSClient.SendSms(&req)
		if err != nil {
			logger.Errorf("failed to send sms: %s", err)
		}
	}()

	return code, nil
}

func (s *SMSProvider) VerifyBindDevice(ctx context.Context, iuser user.Info, code string) (bool, user.Info, error) {
	var cacheUser user.DefaultInfo
	if err := s.cache.Get(ctx, fmt.Sprintf(constant.SMSBindCacheKeyFormat, iuser.GetID(), code), &cacheUser); err != nil {
		if errors.Is(err, cache.ErrNotExists) {
			return false, nil, nil
		}
		logger.Errorf("failed to get user from cache: %s", err)
		return false, nil, err
	}
	go func() {
		if err := s.cache.Remove(context.TODO(), fmt.Sprintf(constant.SMSBindCacheKeyFormat, iuser.GetID(), code)); err != nil {
			logger.Warnf("failed to remove email bind code from cache: %s", err)
		}
	}()
	return true, &cacheUser, nil
}

func (s *SMSProvider) IssueTo(ctx context.Context, user user.Info) (string, error) {
	exist, err := s.cache.Exist(ctx, fmt.Sprintf(constant.SMSVerifyRateLimitKeyFormat, user.GetID()))
	if err != nil {
		logger.Errorf("failed to check rate limit: %s", err)
		return "", err
	}
	if exist {
		return "", errdetails.SendSMSTooFrequently("send sms too frequently, retry after %v sec", s.rateLimitInterval.Seconds())
	}

	code := random.RandDigitString(s.AliyunSMSConfig.CodeLength)

	if err := s.cache.Set(ctx, fmt.Sprintf(constant.SMSVerifyCacheKeyFormat, user.GetID(), code), user, s.expire); err != nil {
		logger.Errorf("failed to cache sms bind code: %s", err)
		return "", err
	}

	go func() {
		if err := s.cache.Set(ctx, fmt.Sprintf(constant.SMSVerifyRateLimitKeyFormat, user.GetID()), "", s.rateLimitInterval); err != nil {
			logger.Errorf("failed to cache email bind rate limit: %s", err)
		}
	}()

	go func() {
		logger.Debug("send sms", zap.String("phone", user.GetPhone()), zap.String("code", code))
		req := dysmsapi.SendSmsRequest{}
		req.SetSignName(s.AliyunSMSConfig.SignName)
		req.SetTemplateCode(s.AliyunSMSConfig.TemplateCode)
		req.SetPhoneNumbers(user.GetPhone())
		req.SetTemplateParam(fmt.Sprintf("{\"code\":\"%s\"}", code))
		_, err := s.aliyunSMSClient.SendSms(&req)
		if err != nil {
			logger.Errorf("failed to send sms: %s", err)
		}
	}()

	return code, nil
}

func (s *SMSProvider) AuthenticationToken(ctx context.Context, iuser user.Info, token string, _ string) (user.Info, error) {
	var cacheUser user.DefaultInfo
	if err := s.cache.Get(ctx, fmt.Sprintf(constant.SMSVerifyCacheKeyFormat, iuser.GetID(), token), &cacheUser); err != nil {
		if errors.Is(err, cache.ErrNotExists) {
			return nil, errdetails.Forbidden("invalid sms verification code")
		}
		logger.Errorf("failed to get user from cache: %s", err)
		return nil, err
	}
	go func() {
		if err := s.cache.Remove(context.TODO(), fmt.Sprintf(constant.SMSVerifyCacheKeyFormat, iuser.GetID(), token)); err != nil {
			logger.Warnf("failed to remove email verification code from cache: %s", err)
		}
	}()
	return &cacheUser, nil
}
