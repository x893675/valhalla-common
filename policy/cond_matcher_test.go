package policy

import (
	"encoding/json"
	"testing"
)

func TestConditionMather(t *testing.T) {
	tests := []struct {
		name           string
		conditionCtx   ConditionContext
		condition      Condition
		expectedResult bool
		expectError    bool
	}{
		{
			name: "IP地址匹配 - 精确匹配",
			conditionCtx: ConditionContext{
				"acs:SourceIp": "10.0.0.1",
			},
			condition: Condition{
				IPAddress: ConditionValue{
					"acs:SourceIp": []string{"10.0.0.1", "192.168.1.1"},
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name: "IP地址匹配 - CIDR匹配",
			conditionCtx: ConditionContext{
				"acs:SourceIp": "192.168.234.50",
			},
			condition: Condition{
				IPAddress: ConditionValue{
					"acs:SourceIp": []string{"192.168.234.0/24"},
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name: "IP地址不匹配",
			conditionCtx: ConditionContext{
				"acs:SourceIp": "127.0.0.1",
			},
			condition: Condition{
				IPAddress: ConditionValue{
					"acs:SourceIp": []string{"10.0.0.1", "192.168.234.0/24"},
				},
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "日期小于 - 匹配",
			conditionCtx: ConditionContext{
				"acs:CurrentTime": "2024-01-10T00:00:00Z",
			},
			condition: Condition{
				DateLessThan: ConditionValue{
					"acs:CurrentTime": []string{"2024-01-12T06:59:00Z"},
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name: "日期小于 - 不匹配",
			conditionCtx: ConditionContext{
				"acs:CurrentTime": "2024-01-15T00:00:00Z",
			},
			condition: Condition{
				DateLessThan: ConditionValue{
					"acs:CurrentTime": []string{"2024-01-12T06:59:00Z"},
				},
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "字符串相等 - 匹配",
			conditionCtx: ConditionContext{
				"acs:UserRole": "admin",
			},
			condition: Condition{
				StringEquals: ConditionValue{
					"acs:UserRole": []string{"admin", "superuser"},
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name: "字符串相等 - 不匹配",
			conditionCtx: ConditionContext{
				"acs:UserRole": "guest",
			},
			condition: Condition{
				StringEquals: ConditionValue{
					"acs:UserRole": []string{"admin", "superuser"},
				},
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "多个条件 - 全部匹配",
			conditionCtx: ConditionContext{
				"acs:SourceIp":    "10.0.0.1",
				"acs:CurrentTime": "2024-01-10T00:00:00Z",
			},
			condition: Condition{
				IPAddress: ConditionValue{
					"acs:SourceIp": []string{"10.0.0.1"},
				},
				DateLessThan: ConditionValue{
					"acs:CurrentTime": []string{"2024-01-12T00:00:00Z"},
				},
			},
			expectedResult: true,
			expectError:    false,
		},
		{
			name: "多个条件 - 部分不匹配",
			conditionCtx: ConditionContext{
				"acs:SourceIp":    "10.0.0.1",
				"acs:CurrentTime": "2024-01-15T00:00:00Z",
			},
			condition: Condition{
				IPAddress: ConditionValue{
					"acs:SourceIp": []string{"10.0.0.1"},
				},
				DateLessThan: ConditionValue{
					"acs:CurrentTime": []string{"2024-01-12T00:00:00Z"},
				},
			},
			expectedResult: false,
			expectError:    false,
		},
		{
			name: "上下文缺少必需字段",
			conditionCtx: ConditionContext{
				"acs:SourceIp": "10.0.0.1",
			},
			condition: Condition{
				DateLessThan: ConditionValue{
					"acs:CurrentTime": []string{"2024-01-12T00:00:00Z"},
				},
			},
			expectedResult: false,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 序列化上下文和条件
			ctxJSON, err := json.Marshal(tt.conditionCtx)
			if err != nil {
				t.Fatalf("Failed to marshal context: %v", err)
			}

			condJSON, err := json.Marshal(tt.condition)
			if err != nil {
				t.Fatalf("Failed to marshal condition: %v", err)
			}

			// 调用 ConditionMather
			result, err := ConditionMather(string(ctxJSON), string(condJSON))

			// 检查错误
			if (err != nil) != tt.expectError {
				t.Errorf("ConditionMather() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// 检查结果
			if !tt.expectError {
				gotResult := result.(bool)
				if gotResult != tt.expectedResult {
					t.Errorf("ConditionMather() = %v, want %v", gotResult, tt.expectedResult)
					t.Logf("Context: %s", string(ctxJSON))
					t.Logf("Condition: %s", string(condJSON))
				}
			}
		})
	}
}

func TestConditionMatherWithInvalidJSON(t *testing.T) {
	tests := []struct {
		name      string
		ctx       string
		condition string
	}{
		{
			name:      "无效的上下文 JSON",
			ctx:       "{invalid json}",
			condition: `{"StringEquals":{"key":["value"]}}`,
		},
		{
			name:      "无效的条件 JSON",
			ctx:       `{"key":"value"}`,
			condition: "{invalid json}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ConditionMather(tt.ctx, tt.condition)
			if err == nil {
				t.Error("ConditionMather() should return error for invalid JSON")
			}
		})
	}
}

func TestConditionMatherWithUnknownOperator(t *testing.T) {
	ctx := ConditionContext{
		"key": "value",
	}
	condition := map[string]ConditionValue{
		"UnknownOperator": {
			"key": []string{"value"},
		},
	}

	ctxJSON, _ := json.Marshal(ctx)
	condJSON, _ := json.Marshal(condition)

	result, err := ConditionMather(string(ctxJSON), string(condJSON))
	if err != nil {
		t.Errorf("ConditionMather() unexpected error: %v", err)
	}

	// 未知操作符应该返回 false
	if result.(bool) != false {
		t.Errorf("ConditionMather() with unknown operator should return false, got %v", result)
	}
}
