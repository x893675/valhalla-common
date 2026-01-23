# ID Generator (idgen)

基于 Sonyflake 的分布式唯一 ID 生成器。

## 特性

- ✨ 基于 Sonyflake 算法，生成唯一的 64 位 ID
- 🔒 线程安全，支持高并发场景
- 🎯 提供多种 ID 格式（uint64、string、带前缀的 string）
- ⚡ 高性能，单次生成耗时约 39µs（Apple M2）
- 🛠️ 支持自定义配置

## 安装

```bash
go get github.com/sony/sonyflake
```

## 使用方法

### 基本使用

```go
import "github.com/x893675/valhalla-common/utils/idgen"

// 生成 uint64 ID
id, err := idgen.NextID()
if err != nil {
    // 处理错误
}

// 生成 uint64 ID (panic on error)
id := idgen.MustNextID()

// 生成字符串 ID
idStr, err := idgen.NextIDString()
if err != nil {
    // 处理错误
}

// 生成字符串 ID (panic on error)
idStr := idgen.MustNextIDString()
```

### 带前缀的 ID

适用于需要区分不同类型资源的场景：

```go
// 生成用户 ID
userID := idgen.MustNextIDStringWithPrefix("user")
// 输出: user:123456789012345

// 生成订单 ID
orderID := idgen.MustNextIDStringWithPrefix("order")
// 输出: order:123456789012346

// 空前缀时只返回 ID
id := idgen.MustNextIDStringWithPrefix("")
// 输出: 123456789012347
```

### 自定义配置

```go
import (
    "time"
    "github.com/sony/sonyflake"
    "github.com/x893675/valhalla-common/utils/idgen"
)

// 在应用启动时自定义配置
func init() {
    idgen.Initialize(sonyflake.Settings{
        StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
        MachineID: func() (uint16, error) {
            // 返回自定义的机器 ID
            return 1, nil
        },
    })
}
```

## API 文档

### 函数列表

| 函数 | 返回值 | 说明 |
|------|--------|------|
| `NextID()` | `(uint64, error)` | 生成下一个唯一 ID |
| `MustNextID()` | `uint64` | 生成下一个唯一 ID，出错时 panic |
| `NextIDString()` | `(string, error)` | 生成下一个唯一 ID 的字符串形式 |
| `MustNextIDString()` | `string` | 生成下一个唯一 ID 的字符串形式，出错时 panic |
| `NextIDStringWithPrefix(prefix string)` | `(string, error)` | 生成带前缀的 ID 字符串 |
| `MustNextIDStringWithPrefix(prefix string)` | `string` | 生成带前缀的 ID 字符串，出错时 panic |
| `Initialize(settings sonyflake.Settings)` | `void` | 初始化 ID 生成器（可选，使用 sync.Once 保证只执行一次） |

## 性能指标

在 Apple M2 处理器上的性能测试结果：

| 操作 | 耗时 | 内存分配 | 分配次数 |
|------|------|----------|----------|
| NextID | ~39µs | 0 B | 0 |
| MustNextID | ~39µs | 0 B | 0 |
| NextIDString | ~39µs | 24 B | 1 |
| MustNextIDString | ~39µs | 24 B | 1 |
| NextIDStringWithPrefix (短前缀) | ~39µs | 80 B | 4 |
| NextIDStringWithPrefix (长前缀) | ~39µs | 120 B | 4 |

**并发性能**：
- 单线程：~39µs/op
- 并发场景：~39µs/op (性能保持稳定)
- 1000 协程并发：~38µs/op

## 测试

### 运行单元测试

```bash
go test -v ./utils/idgen/
```

### 运行性能测试

```bash
go test -bench=. -benchmem ./utils/idgen/
```

### 测试覆盖率

```bash
go test -cover ./utils/idgen/
```

## 优化说明

相比原实现，本次优化包括：

### 1. 修复了 Bug
- **修复前**: `MustNextIDStringWithPrefix` 在错误时返回空字符串 `""`
- **修复后**: 出错时 panic，与其他 `Must*` 函数行为一致

### 2. 改进了初始化逻辑
- **修复前**: 使用全局变量直接初始化，无法自定义配置
- **修复后**: 使用 `sync.Once` 实现延迟初始化，支持自定义配置

### 3. 增强了错误处理
- 所有 `Must*` 函数在 panic 时提供更详细的错误信息
- 错误信息使用 `fmt.Errorf` 包装，便于错误追踪

### 4. 新增了函数
- `NextIDStringWithPrefix()`: 非 panic 版本的带前缀 ID 生成
- `Initialize()`: 允许自定义 sonyflake 配置

### 5. 完善的测试
- ✅ 单元测试：100% 覆盖所有函数
- ✅ 并发测试：验证高并发场景下的正确性
- ✅ 性能测试：全面的性能基准测试
- ✅ 示例代码：提供实用的使用示例

## 注意事项

1. **线程安全**: 所有函数都是线程安全的，可以在多个 goroutine 中并发调用
2. **ID 递增性**: 生成的 ID 是严格递增的（在同一进程内）
3. **唯一性保证**: 只要机器 ID 不同，即使在分布式环境下也能保证唯一性
4. **配置一次**: `Initialize()` 使用 `sync.Once` 实现，多次调用只有第一次有效
5. **性能考虑**: 
   - `NextID()` 性能最佳（无内存分配）
   - `NextIDString()` 有 1 次内存分配
   - `NextIDStringWithPrefix()` 有 4 次内存分配

## 关于 Sonyflake

Sonyflake 是 Sony 开发的分布式唯一 ID 生成算法，类似 Twitter Snowflake，但具有：

- 更长的生存周期（174 年 vs 69 年）
- 更多的序列号位（8 位 vs 12 位）
- 更少的机器 ID 位（16 位 vs 10 位）
- 39 位时间戳（10ms 精度）

ID 结构：`[39位时间戳][8位序列号][16位机器ID]`

## License

与项目主 License 保持一致。
