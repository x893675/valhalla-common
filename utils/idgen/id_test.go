package idgen

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/sony/sonyflake"
)

func TestNextID(t *testing.T) {
	id1, err := NextID()
	if err != nil {
		t.Fatalf("NextID() error = %v", err)
	}
	if id1 == 0 {
		t.Error("NextID() returned 0")
	}

	id2, err := NextID()
	if err != nil {
		t.Fatalf("NextID() error = %v", err)
	}
	if id2 == 0 {
		t.Error("NextID() returned 0")
	}

	if id1 == id2 {
		t.Error("NextID() returned duplicate IDs")
	}

	if id2 <= id1 {
		t.Error("NextID() returned non-increasing IDs")
	}
}

func TestMustNextID(t *testing.T) {
	id1 := MustNextID()
	if id1 == 0 {
		t.Error("MustNextID() returned 0")
	}

	id2 := MustNextID()
	if id2 == 0 {
		t.Error("MustNextID() returned 0")
	}

	if id1 == id2 {
		t.Error("MustNextID() returned duplicate IDs")
	}

	if id2 <= id1 {
		t.Error("MustNextID() returned non-increasing IDs")
	}
}

func TestNextIDString(t *testing.T) {
	id1, err := NextIDString()
	if err != nil {
		t.Fatalf("NextIDString() error = %v", err)
	}
	if id1 == "" || id1 == "0" {
		t.Error("NextIDString() returned invalid ID")
	}

	// 验证返回的是有效的数字字符串
	if _, err := strconv.ParseUint(id1, 10, 64); err != nil {
		t.Errorf("NextIDString() returned invalid number string: %v", err)
	}

	id2, err := NextIDString()
	if err != nil {
		t.Fatalf("NextIDString() error = %v", err)
	}

	if id1 == id2 {
		t.Error("NextIDString() returned duplicate IDs")
	}
}

func TestMustNextIDString(t *testing.T) {
	id1 := MustNextIDString()
	if id1 == "" || id1 == "0" {
		t.Error("MustNextIDString() returned invalid ID")
	}

	// 验证返回的是有效的数字字符串
	if _, err := strconv.ParseUint(id1, 10, 64); err != nil {
		t.Errorf("MustNextIDString() returned invalid number string: %v", err)
	}

	id2 := MustNextIDString()
	if id1 == id2 {
		t.Error("MustNextIDString() returned duplicate IDs")
	}
}

func TestNextIDStringWithPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{
			name:   "with prefix",
			prefix: "user",
		},
		{
			name:   "empty prefix",
			prefix: "",
		},
		{
			name:   "complex prefix",
			prefix: "tenant_123_user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := NextIDStringWithPrefix(tt.prefix)
			if err != nil {
				t.Fatalf("NextIDStringWithPrefix() error = %v", err)
			}

			if tt.prefix != "" {
				expected := tt.prefix + "-"
				if !strings.HasPrefix(id, expected) {
					t.Errorf("NextIDStringWithPrefix() = %v, want prefix %v", id, expected)
				}

				// 验证前缀后面是有效的 ID
				parts := strings.SplitN(id, "-", 2)
				if len(parts) != 2 {
					t.Errorf("NextIDStringWithPrefix() returned invalid format: %v", id)
				}
				if _, err := strconv.ParseUint(parts[1], 10, 64); err != nil {
					t.Errorf("NextIDStringWithPrefix() returned invalid ID part: %v", err)
				}
			} else {
				// 空前缀时应该只返回 ID
				if _, err := strconv.ParseUint(id, 10, 64); err != nil {
					t.Errorf("NextIDStringWithPrefix() with empty prefix returned invalid ID: %v", err)
				}
			}
		})
	}
}

func TestMustNextIDStringWithPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{
			name:   "with prefix",
			prefix: "order",
		},
		{
			name:   "empty prefix",
			prefix: "",
		},
		{
			name:   "complex prefix",
			prefix: "workspace_456_resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := MustNextIDStringWithPrefix(tt.prefix)

			if tt.prefix != "" {
				expected := tt.prefix + "-"
				if !strings.HasPrefix(id, expected) {
					t.Errorf("MustNextIDStringWithPrefix() = %v, want prefix %v", id, expected)
				}

				// 验证前缀后面是有效的 ID
				parts := strings.SplitN(id, "-", 2)
				if len(parts) != 2 {
					t.Errorf("MustNextIDStringWithPrefix() returned invalid format: %v", id)
				}
				if _, err := strconv.ParseUint(parts[1], 10, 64); err != nil {
					t.Errorf("MustNextIDStringWithPrefix() returned invalid ID part: %v", err)
				}
			} else {
				// 空前缀时应该只返回 ID
				if _, err := strconv.ParseUint(id, 10, 64); err != nil {
					t.Errorf("MustNextIDStringWithPrefix() with empty prefix returned invalid ID: %v", err)
				}
			}
		})
	}
}

func TestInitialize(t *testing.T) {
	// 注意：这个测试需要在独立的进程中运行，因为 sync.Once 只会执行一次
	// 这里我们测试多次调用 Initialize 不会 panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Initialize() should not panic on multiple calls, got: %v", r)
		}
	}()

	// 第一次调用
	Initialize(sonyflake.Settings{})
	id1 := MustNextID()

	// 第二次调用（应该被忽略）
	Initialize(sonyflake.Settings{})
	id2 := MustNextID()

	if id1 >= id2 {
		t.Error("IDs should be increasing")
	}
}

// TestConcurrentNextID 测试并发生成 ID 的正确性
func TestConcurrentNextID(t *testing.T) {
	const (
		goroutines = 100
		idsPerGo   = 100
	)

	idMap := sync.Map{}
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGo; j++ {
				id := MustNextID()
				if _, loaded := idMap.LoadOrStore(id, true); loaded {
					t.Errorf("Duplicate ID generated: %d", id)
				}
			}
		}()
	}

	wg.Wait()

	// 验证生成了正确数量的唯一 ID
	count := 0
	idMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	expected := goroutines * idsPerGo
	if count != expected {
		t.Errorf("Expected %d unique IDs, got %d", expected, count)
	}
}

// TestConcurrentNextIDString 测试并发生成字符串 ID 的正确性
func TestConcurrentNextIDString(t *testing.T) {
	const (
		goroutines = 100
		idsPerGo   = 100
	)

	idMap := sync.Map{}
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < idsPerGo; j++ {
				id := MustNextIDString()
				if _, loaded := idMap.LoadOrStore(id, true); loaded {
					t.Errorf("Duplicate ID generated: %s", id)
				}
			}
		}()
	}

	wg.Wait()

	// 验证生成了正确数量的唯一 ID
	count := 0
	idMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	expected := goroutines * idsPerGo
	if count != expected {
		t.Errorf("Expected %d unique IDs, got %d", expected, count)
	}
}

// TestConcurrentNextIDStringWithPrefix 测试并发生成带前缀字符串 ID 的正确性
func TestConcurrentNextIDStringWithPrefix(t *testing.T) {
	const (
		goroutines = 50
		idsPerGo   = 100
	)

	prefixes := []string{"user", "order", "product", "tenant", ""}

	idMap := sync.Map{}
	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		prefix := prefixes[i%len(prefixes)]
		go func(p string) {
			defer wg.Done()
			for j := 0; j < idsPerGo; j++ {
				id := MustNextIDStringWithPrefix(p)
				if _, loaded := idMap.LoadOrStore(id, true); loaded {
					t.Errorf("Duplicate ID generated: %s", id)
				}
			}
		}(prefix)
	}

	wg.Wait()

	// 验证生成了正确数量的唯一 ID
	count := 0
	idMap.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	expected := goroutines * idsPerGo
	if count != expected {
		t.Errorf("Expected %d unique IDs, got %d", expected, count)
	}
}

// TestIDOrdering 测试 ID 的单调递增性
func TestIDOrdering(t *testing.T) {
	const count = 1000
	ids := make([]uint64, count)

	for i := 0; i < count; i++ {
		ids[i] = MustNextID()
	}

	for i := 1; i < count; i++ {
		if ids[i] <= ids[i-1] {
			t.Errorf("IDs are not strictly increasing: ids[%d]=%d, ids[%d]=%d",
				i-1, ids[i-1], i, ids[i])
		}
	}
}

// 示例：基本使用
func ExampleNextID() {
	id, err := NextID()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Generated ID: %d\n", id)
}

// 示例：使用 Must 版本
func ExampleMustNextID() {
	id := MustNextID()
	fmt.Printf("Generated ID: %d\n", id)
}

// 示例：生成字符串 ID
func ExampleNextIDString() {
	id, err := NextIDString()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Generated ID: %s\n", id)
}

// 示例：生成带前缀的 ID
func ExampleMustNextIDStringWithPrefix() {
	userID := MustNextIDStringWithPrefix("user")
	orderID := MustNextIDStringWithPrefix("order")
	fmt.Printf("User ID: %s\n", userID)
	fmt.Printf("Order ID: %s\n", orderID)
}
