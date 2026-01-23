package idgen

import (
	"sync"
	"testing"
)

// BenchmarkNextID 测试 NextID 的性能
func BenchmarkNextID(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NextID()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMustNextID 测试 MustNextID 的性能
func BenchmarkMustNextID(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MustNextID()
	}
}

// BenchmarkNextIDString 测试 NextIDString 的性能
func BenchmarkNextIDString(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NextIDString()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkMustNextIDString 测试 MustNextIDString 的性能
func BenchmarkMustNextIDString(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MustNextIDString()
	}
}

// BenchmarkNextIDStringWithPrefix 测试 NextIDStringWithPrefix 的性能
func BenchmarkNextIDStringWithPrefix(b *testing.B) {
	benchmarks := []struct {
		name   string
		prefix string
	}{
		{"EmptyPrefix", ""},
		{"ShortPrefix", "user"},
		{"LongPrefix", "tenant_12345_workspace_67890_resource"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := NextIDStringWithPrefix(bm.prefix)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkMustNextIDStringWithPrefix 测试 MustNextIDStringWithPrefix 的性能
func BenchmarkMustNextIDStringWithPrefix(b *testing.B) {
	benchmarks := []struct {
		name   string
		prefix string
	}{
		{"EmptyPrefix", ""},
		{"ShortPrefix", "user"},
		{"LongPrefix", "tenant_12345_workspace_67890_resource"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = MustNextIDStringWithPrefix(bm.prefix)
			}
		})
	}
}

// BenchmarkNextIDParallel 测试并发场景下的 NextID 性能
func BenchmarkNextIDParallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := NextID()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkMustNextIDParallel 测试并发场景下的 MustNextID 性能
func BenchmarkMustNextIDParallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = MustNextID()
		}
	})
}

// BenchmarkNextIDStringParallel 测试并发场景下的 NextIDString 性能
func BenchmarkNextIDStringParallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := NextIDString()
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkMustNextIDStringParallel 测试并发场景下的 MustNextIDString 性能
func BenchmarkMustNextIDStringParallel(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = MustNextIDString()
		}
	})
}

// BenchmarkNextIDStringWithPrefixParallel 测试并发场景下的 NextIDStringWithPrefix 性能
func BenchmarkNextIDStringWithPrefixParallel(b *testing.B) {
	benchmarks := []struct {
		name   string
		prefix string
	}{
		{"EmptyPrefix", ""},
		{"ShortPrefix", "user"},
		{"LongPrefix", "tenant_12345_workspace_67890_resource"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, err := NextIDStringWithPrefix(bm.prefix)
					if err != nil {
						b.Fatal(err)
					}
				}
			})
		})
	}
}

// BenchmarkMustNextIDStringWithPrefixParallel 测试并发场景下的 MustNextIDStringWithPrefix 性能
func BenchmarkMustNextIDStringWithPrefixParallel(b *testing.B) {
	benchmarks := []struct {
		name   string
		prefix string
	}{
		{"EmptyPrefix", ""},
		{"ShortPrefix", "user"},
		{"LongPrefix", "tenant_12345_workspace_67890_resource"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_ = MustNextIDStringWithPrefix(bm.prefix)
				}
			})
		})
	}
}

// BenchmarkConcurrentLoad 模拟真实场景的高并发负载
func BenchmarkConcurrentLoad(b *testing.B) {
	scenarios := []struct {
		name       string
		goroutines int
	}{
		{"10Goroutines", 10},
		{"100Goroutines", 100},
		{"1000Goroutines", 1000},
	}

	for _, scenario := range scenarios {
		b.Run(scenario.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()

			var wg sync.WaitGroup
			opsPerGoroutine := b.N / scenario.goroutines
			if opsPerGoroutine == 0 {
				opsPerGoroutine = 1
			}

			for i := 0; i < scenario.goroutines; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for j := 0; j < opsPerGoroutine; j++ {
						_ = MustNextID()
					}
				}()
			}
			wg.Wait()
		})
	}
}

// BenchmarkMixedOperations 测试混合操作的性能
func BenchmarkMixedOperations(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			switch i % 4 {
			case 0:
				_ = MustNextID()
			case 1:
				_ = MustNextIDString()
			case 2:
				_ = MustNextIDStringWithPrefix("user")
			case 3:
				_ = MustNextIDStringWithPrefix("")
			}
			i++
		}
	})
}

// BenchmarkCompareImplementations 比较不同实现方式的性能
func BenchmarkCompareImplementations(b *testing.B) {
	b.Run("DirectID", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = MustNextID()
		}
	})

	b.Run("IDToString", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = MustNextIDString()
		}
	})

	b.Run("IDWithPrefix", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = MustNextIDStringWithPrefix("user")
		}
	})
}
