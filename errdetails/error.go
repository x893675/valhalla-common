/*
Copyright 2024 x893675.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package errdetails

import (
	"errors"
	"fmt"
	"net/http"
)

// BizError
// 业务自定义 error
type BizError struct {
	HTTPStatusCode int `json:"-"`
	// Code 是自定义错误码
	Code int `json:"code,omitempty" example:"400"`
	// Reason 是具体的错误原因
	Reason  string `json:"reason,omitempty" example:"Bad Request"`
	Message string `json:"message,omitempty" example:"Bad Request"`
	// Metadata 是错误携带的元数据，在错误中可以填入一些自定义字段来保存出现错误时的上下文信息
	Metadata map[string]string `json:"metadata,omitempty" example:"user_id:workflowgroup"`
	// cause underlying cause of the error
	cause error
}

func (e *BizError) Error() string {
	return fmt.Sprintf("error: code = %d reason = %s message = %s metadata = %v cause = %v", e.Code, e.Reason, e.Message, e.Metadata, e.cause)
}

func (e *BizError) Unwrap() error {
	return e.cause
}

func (e *BizError) Is(err error) bool {
	if se := new(BizError); errors.As(err, &se) {
		return se.Code == e.Code && se.Reason == e.Reason
	}
	return false
}

func (e *BizError) WithCause(err error) *BizError {
	newErr := Clone(e)
	newErr.cause = err
	return newErr
}

func (e *BizError) WithMetadata(md map[string]string) *BizError {
	err := Clone(e)
	err.Metadata = md
	return err
}

func HTTPStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	return FromError(err).HTTPStatusCode
}

func Code(err error) int {
	if err == nil {
		return NoErrorCode
	}
	return FromError(err).Code
}

func Reason(err error) string {
	if err == nil {
		return UnknownReason
	}
	return FromError(err).Reason
}

func Clone(err *BizError) *BizError {
	if err == nil {
		return nil
	}
	metadata := make(map[string]string, len(err.Metadata))
	for k, v := range err.Metadata {
		metadata[k] = v
	}
	return &BizError{
		HTTPStatusCode: err.HTTPStatusCode,
		cause:          err.cause,
		Code:           err.Code,
		Reason:         err.Reason,
		Message:        err.Message,
		Metadata:       metadata,
	}
}

func FromError(err error) *BizError {
	if err == nil {
		return nil
	}
	if se := new(BizError); errors.As(err, &se) {
		return se
	}
	return New(http.StatusInternalServerError, UnknownCode, UnknownReason, err.Error())
}

func New(httpStatusCode, code int, reason, message string) *BizError {
	return &BizError{
		HTTPStatusCode: httpStatusCode,
		Code:           code,
		Reason:         reason,
		Message:        message,
	}
}

func Newf(httpStatusCode, code int, reason, format string, a ...interface{}) *BizError {
	return New(httpStatusCode, code, reason, fmt.Sprintf(format, a...))
}

func Errorf(httpStatusCode, code int, reason, format string, a ...interface{}) error {
	return New(httpStatusCode, code, reason, fmt.Sprintf(format, a...))
}
