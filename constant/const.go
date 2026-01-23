package constant

import "time"

const (
	TOTPURLKey    = "totp_url"
	TOTPSecretKey = "totp_secret"
)

const (
	SecurityLevelNone uint = iota
	SecurityLevelLow
	SecurityLevelHigh
)

const (
	// TODO: make this configurable
	DefaultSessionExpireHours = 12
	MFATokenCacheDuration     = 10 * time.Minute
)

const (
	EnforceUseMFA uint = iota + 1
	EnforceUseMFAForUserSetting
	EnforceUseMFAForAbnormalLogin
)

const (
	MFAProviderTOTP  = "TOTP"
	MFAProviderSMS   = "SMS"
	MFAProviderEmail = "Email"
)

const (
	TOTPCacheKeyPrefix = "totp:"
	TOTPCacheKeyFormat = TOTPCacheKeyPrefix + "%s"

	// EmailBindCacheKeyPrefix
	// 验证邮箱时的缓存key，  email-bind:uid:code: user-info
	EmailBindCacheKeyPrefix = "email-bind:"
	EmailBindCacheKeyFormat = EmailBindCacheKeyPrefix + "%s:" + "%s"

	// EmailVerifyCacheKeyPrefix
	// 发送邮件验证码时的缓存key，  email-code:uid:code
	EmailVerifyCacheKeyPrefix = "email-code:"
	EmailVerifyCacheKeyFormat = EmailVerifyCacheKeyPrefix + "%s:%s"

	// SMSBindCacheKeyPrefix
	// 验证手机号的缓存key
	// 验证手机号时的缓存key，  sms-bind:uid:code: user-info
	SMSBindCacheKeyPrefix     = "sms-bind:"
	SMSBindCacheKeyFormat     = SMSBindCacheKeyPrefix + "%s:" + "%s"
	SMSBindRateLimitKeyFormat = SMSBindCacheKeyPrefix + "rate-limit:%s"

	// SMSVerifyCacheKeyPrefix
	// 发送短信验证码时的缓存key，  sms-code:uid:code
	SMSVerifyCacheKeyPrefix     = "sms-code:"
	SMSVerifyCacheKeyFormat     = SMSVerifyCacheKeyPrefix + "%s:%s"
	SMSVerifyRateLimitKeyFormat = SMSVerifyCacheKeyPrefix + "rate-limit:%s"

	// TokenCacheKeyPrefix
	// cache key pattern: token:<uid>:<token_str>:<user.info>
	TokenCacheKeyPrefix = "token:%s:"
	TokenCacheKeyFormat = TokenCacheKeyPrefix + "%s"

	MFAVerifyCacheKeyPrefix = "mfa-verify:"
	MFAVerifyCacheKeyFormat = MFAVerifyCacheKeyPrefix + "%s"

	MFALoginCacheKeyPrefix = "mfa-login:"
	MFALoginCacheKeyFormat = MFALoginCacheKeyPrefix + "%s"
)
