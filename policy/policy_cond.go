package policy

import (
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	StringEquals              = "StringEquals"
	StringNotEquals           = "StringNotEquals"
	StringEqualsIgnoreCase    = "StringEqualsIgnoreCase"
	StringNotEqualsIgnoreCase = "StringNotEqualsIgnoreCase"
	StringLike                = "StringLike"
	StringNotLike             = "StringNotLike"

	NumericEquals            = "NumericEquals"
	NumericNotEquals         = "NumericNotEquals"
	NumericLessThan          = "NumericLessThan"
	NumericLessThanEquals    = "NumericLessThanEquals"
	NumericGreaterThan       = "NumericGreaterThan"
	NumericGreaterThanEquals = "NumericGreaterThanEquals"

	DateEquals            = "DateEquals"
	DateNotEquals         = "DateNotEquals"
	DateLessThan          = "DateLessThan"
	DateLessThanEquals    = "DateLessThanEquals"
	DateGreaterThan       = "DateGreaterThan"
	DateGreaterThanEquals = "DateGreaterThanEquals"

	Bool = "Bool"

	IPAddress    = "IPAddress"
	NotIPAddress = "NotIPAddress"
)

type ConditionOperatorFunc func(param1, param2 interface{}) bool

var conditionOperatorFuncMap = map[string]ConditionOperatorFunc{
	StringEquals:              StringEqualsFunc,
	StringNotEquals:           StringNotEqualsFunc,
	StringEqualsIgnoreCase:    StringEqualsIgnoreCaseFunc,
	StringNotEqualsIgnoreCase: StringNotEqualsIgnoreCaseFunc,
	StringLike:                StringLikeFunc,
	StringNotLike:             StringNotLikeFunc,
	NumericEquals:             NumericEqualsFunc,
	NumericNotEquals:          NumericNotEqualsFunc,
	NumericLessThan:           NumericLessThanFunc,
	NumericLessThanEquals:     NumericLessThanEqualsFunc,
	NumericGreaterThan:        NumericGreaterThanFunc,
	NumericGreaterThanEquals:  NumericGreaterThanEqualsFunc,
	DateEquals:                DateEqualsFunc,
	DateNotEquals:             DateNotEqualsFunc,
	DateLessThan:              DateLessThanFunc,
	DateLessThanEquals:        DateLessThanEqualsFunc,
	DateGreaterThan:           DateGreaterThanFunc,
	DateGreaterThanEquals:     DateGreaterThanEqualsFunc,
	Bool:                      BoolFunc,
	IPAddress:                 IPAddressFunc,
	NotIPAddress:              NotIPAddressFunc,
}

// 泛型辅助函数：对列表中的任意元素进行匹配
// compareFn 定义具体的比较逻辑
func anyMatch[T any](value T, values []T, compareFn func(T, T) bool) bool {
	for _, v := range values {
		if compareFn(value, v) {
			return true
		}
	}
	return false
}

// 泛型相等比较
func equals[T comparable](value T, values []T) bool {
	return anyMatch(value, values, func(a, b T) bool {
		return a == b
	})
}

// 泛型不等比较
func notEquals[T comparable](value T, values []T) bool {
	return anyMatch(value, values, func(a, b T) bool {
		return a != b
	})
}

// 泛型有序类型约束
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string
}

// 泛型小于比较
func lessThan[T Ordered](value T, values []T) bool {
	return anyMatch(value, values, func(a, b T) bool {
		return a < b
	})
}

// 泛型小于等于比较
func lessThanEquals[T Ordered](value T, values []T) bool {
	return anyMatch(value, values, func(a, b T) bool {
		return a <= b
	})
}

// 泛型大于比较
func greaterThan[T Ordered](value T, values []T) bool {
	return anyMatch(value, values, func(a, b T) bool {
		return a > b
	})
}

// 泛型大于等于比较
func greaterThanEquals[T Ordered](value T, values []T) bool {
	return anyMatch(value, values, func(a, b T) bool {
		return a >= b
	})
}

// 字符串比较函数
func StringEqualsFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)
	return equals(value, values)
}

func StringNotEqualsFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)
	return notEquals(value, values)
}

func StringEqualsIgnoreCaseFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)
	return anyMatch(value, values, func(a, b string) bool {
		return strings.EqualFold(a, b)
	})
}

func StringNotEqualsIgnoreCaseFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)
	return anyMatch(value, values, func(a, b string) bool {
		return !strings.EqualFold(a, b)
	})
}

func StringLikeFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)
	return anyMatch(value, values, func(a, b string) bool {
		return strings.Contains(a, b)
	})
}

func StringNotLikeFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)
	return anyMatch(value, values, func(a, b string) bool {
		return !strings.Contains(a, b)
	})
}

// 数值比较函数
func NumericEqualsFunc(param1, param2 interface{}) bool {
	value := param1.(int)
	values := param2.([]int)
	return equals(value, values)
}

func NumericNotEqualsFunc(param1, param2 interface{}) bool {
	value := param1.(int)
	values := param2.([]int)
	return notEquals(value, values)
}

func NumericLessThanFunc(param1, param2 interface{}) bool {
	value := param1.(int)
	values := param2.([]int)
	return lessThan(value, values)
}

func NumericLessThanEqualsFunc(param1, param2 interface{}) bool {
	value := param1.(int)
	values := param2.([]int)
	return lessThanEquals(value, values)
}

func NumericGreaterThanFunc(param1, param2 interface{}) bool {
	value := param1.(int)
	values := param2.([]int)
	return greaterThan(value, values)
}

func NumericGreaterThanEqualsFunc(param1, param2 interface{}) bool {
	value := param1.(int)
	values := param2.([]int)
	return greaterThanEquals(value, values)
}

// 日期比较函数
func DateEqualsFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)
	return anyMatch(value, values, func(a, b string) bool {
		aTime, _ := time.Parse(time.RFC3339, a)
		bTime, _ := time.Parse(time.RFC3339, b)
		return aTime.Equal(bTime)
	})
}

func DateNotEqualsFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)
	return anyMatch(value, values, func(a, b string) bool {
		aTime, _ := time.Parse(time.RFC3339, a)
		bTime, _ := time.Parse(time.RFC3339, b)
		return !aTime.Equal(bTime)
	})
}

func DateLessThanFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)
	return anyMatch(value, values, func(a, b string) bool {
		aTime, _ := time.Parse(time.RFC3339, a)
		bTime, _ := time.Parse(time.RFC3339, b)
		return aTime.Before(bTime)
	})
}

func DateLessThanEqualsFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)
	return anyMatch(value, values, func(a, b string) bool {
		aTime, _ := time.Parse(time.RFC3339, a)
		bTime, _ := time.Parse(time.RFC3339, b)
		return aTime.Before(bTime) || aTime.Equal(bTime)
	})
}

func DateGreaterThanFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)
	return anyMatch(value, values, func(a, b string) bool {
		aTime, _ := time.Parse(time.RFC3339, a)
		bTime, _ := time.Parse(time.RFC3339, b)
		return aTime.After(bTime)
	})
}

func DateGreaterThanEqualsFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)
	return anyMatch(value, values, func(a, b string) bool {
		aTime, _ := time.Parse(time.RFC3339, a)
		bTime, _ := time.Parse(time.RFC3339, b)
		return aTime.After(bTime) || aTime.Equal(bTime)
	})
}

// 布尔值比较函数
func BoolFunc(param1, param2 interface{}) bool {
	value := param1.(bool)
	values := param2.([]bool)
	return equals(value, values)
}

// IP 地址比较函数
func IPAddressFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)

	requestIP := net.ParseIP(value)
	if requestIP == nil {
		return false
	}

	return anyMatch(value, values, func(_, policyValue string) bool {
		policyIP := net.ParseIP(policyValue)
		if policyIP == nil {
			// 尝试解析为 CIDR
			_, policyNet, err := net.ParseCIDR(policyValue)
			if err != nil {
				return false
			}
			return policyNet.Contains(requestIP)
		}
		return requestIP.Equal(policyIP)
	})
}

func NotIPAddressFunc(param1, param2 interface{}) bool {
	value := param1.(string)
	values := param2.([]string)

	requestIP := net.ParseIP(value)
	if requestIP == nil {
		return false
	}

	return anyMatch(value, values, func(_, policyValue string) bool {
		policyIP := net.ParseIP(policyValue)
		if policyIP == nil {
			// 尝试解析为 CIDR
			_, policyNet, err := net.ParseCIDR(policyValue)
			if err != nil {
				return false
			}
			return !policyNet.Contains(requestIP)
		}
		return !requestIP.Equal(policyIP)
	})
}

type ConditionParser interface {
	ParseCondition(req *http.Request) any
}

var ConditionKeyMap = map[string]ConditionParser{
	"inf:SourceIP":    &SourceIP{},
	"inf:CurrentTime": &CurrentTime{},
	"iam:ServiceName": &Service{},
}

type ConditionContext map[string]any
