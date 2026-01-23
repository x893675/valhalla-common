package errdetails

import (
	"fmt"
	"net/http"
)

const (
	NoErrorCode          = 200
	InvalidParameterCode = 400
	UnauthorizedCode     = 401
	FobiddenCode         = 403
	UnknownCode          = 500

	BindParameterFailedCode = 10001
	UnexpectedErrorCode     = 10002

	DatabaseOperationFailedCode = 20000
	ResourceAlreadyExistsCode   = 20001
	ResourceNotFoundCode        = 20002
	CacheOperationFailedCode    = 20003
	RequirePreconditionCode     = 20004
	SendSMSTooFrequentlyCode    = 20005

	NotImplemented = 30000
)

const (
	UnknownReason          = "Unknown"
	UnauthorizedReason     = "Unauthorized"
	ForbiddenReason        = "Forbidden"
	InvalidParameterReason = "InvalidParameter"

	BindParameterFailedReason = "BindParameterFailed"
	UnexpectedErrorReason     = "UnexpectedError"

	DatabaseOperationFailedReason = "DatabaseOperationFailed"
	ResourceAlreadyExistsReason   = "ResourceAlreadyExists"
	ResourceNotFoundReason        = "ResourceNotFound"
	CacheOperationFailedReason    = "CacheOperationFailed"
	RequirePreconditionReason     = "RequirePrecondition"
	SendSMSTooFrequentlyReason    = "SendSMSTooFrequently"

	NotImplementedReason = "NotImplemented"
)

func Unauthorized(format string, a ...interface{}) *BizError {
	return New(http.StatusUnauthorized, UnauthorizedCode, UnauthorizedReason, fmt.Sprintf(format, a...))
}

func IsUnauthorized(err error) bool {
	e := FromError(err)
	return e.Code == UnauthorizedCode && e.Reason == UnauthorizedReason
}

func Forbidden(format string, a ...interface{}) *BizError {
	return New(http.StatusForbidden, FobiddenCode, ForbiddenReason, fmt.Sprintf(format, a...))
}

func IsForbidden(err error) bool {
	e := FromError(err)
	return e.Code == FobiddenCode && e.Reason == ForbiddenReason
}

func BindParameterFailed(format string, a ...interface{}) *BizError {
	return New(http.StatusBadRequest, BindParameterFailedCode, BindParameterFailedReason, fmt.Sprintf(format, a...))
}

func IsBindParameterFailed(err error) bool {
	e := FromError(err)
	return e.Code == BindParameterFailedCode && e.Reason == BindParameterFailedReason
}

func InvalidParameter(format string, a ...interface{}) *BizError {
	return New(http.StatusBadRequest, InvalidParameterCode, InvalidParameterReason, fmt.Sprintf(format, a...))
}

func IsInvalidParameter(err error) bool {
	e := FromError(err)
	return e.Code == InvalidParameterCode && e.Reason == InvalidParameterReason
}

func ResourceAlreadyExists(format string, a ...interface{}) *BizError {
	return New(http.StatusConflict, ResourceAlreadyExistsCode, ResourceAlreadyExistsReason, fmt.Sprintf(format, a...))
}

func IsResourceAlreadyExists(err error) bool {
	e := FromError(err)
	return e.Code == ResourceAlreadyExistsCode && e.Reason == ResourceAlreadyExistsReason
}

func ResourceNotFound(format string, a ...interface{}) *BizError {
	return New(http.StatusNotFound, ResourceNotFoundCode, ResourceNotFoundReason, fmt.Sprintf(format, a...))
}

func IsResourceNotFound(err error) bool {
	e := FromError(err)
	return e.Code == ResourceNotFoundCode && e.Reason == ResourceNotFoundReason
}

func UnexpectedError(format string, a ...interface{}) *BizError {
	return New(http.StatusInternalServerError, UnexpectedErrorCode, UnexpectedErrorReason, fmt.Sprintf(format, a...))
}

func IsUnexpectedError(err error) bool {
	e := FromError(err)
	return e.Code == UnexpectedErrorCode && e.Reason == UnexpectedErrorReason
}

func DatabaseOperationFailed(format string, a ...interface{}) *BizError {
	return New(http.StatusInternalServerError, DatabaseOperationFailedCode, DatabaseOperationFailedReason, fmt.Sprintf(format, a...))
}

func IsDatabaseOperationFailed(err error) bool {
	e := FromError(err)
	return e.Code == DatabaseOperationFailedCode && e.Reason == DatabaseOperationFailedReason
}

func CacheOperationFailed(format string, a ...interface{}) *BizError {
	return New(http.StatusInternalServerError, CacheOperationFailedCode, CacheOperationFailedReason, fmt.Sprintf(format, a...))
}

func IsCacheOperationFailed(err error) bool {
	e := FromError(err)
	return e.Code == CacheOperationFailedCode && e.Reason == CacheOperationFailedReason
}

func NotImplementedError(format string, a ...interface{}) *BizError {
	return New(http.StatusNotImplemented, NotImplemented, NotImplementedReason, fmt.Sprintf(format, a...))
}

func IsNotImplementedError(err error) bool {
	e := FromError(err)
	return e.Code == NotImplemented && e.Reason == NotImplementedReason
}

func SendSMSTooFrequently(format string, a ...interface{}) *BizError {
	return New(http.StatusTooManyRequests, SendSMSTooFrequentlyCode, SendSMSTooFrequentlyReason, fmt.Sprintf(format, a...))
}

func IsSendSMSTooFrequently(err error) bool {
	e := FromError(err)
	return e.Code == SendSMSTooFrequentlyCode && e.Reason == SendSMSTooFrequentlyReason
}

func RequirePrecondition(format string, a ...interface{}) *BizError {
	return New(http.StatusPreconditionRequired, RequirePreconditionCode, RequirePreconditionReason, fmt.Sprintf(format, a...))
}

func IsRequirePrecondition(err error) bool {
	e := FromError(err)
	return e.Code == RequirePreconditionCode && e.Reason == RequirePreconditionReason
}
