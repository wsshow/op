# OP - Go 实用工具包集合

[![Go 版本](https://img.shields.io/badge/Go-1.24+-blue.svg)](https://go.dev)
[![许可证](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

[English](./README.md) | 简体中文

`op` 是一个精心设计的 Go 工具包，提供可复用的、泛型优先的通用编程工具。每个包都追求性能与易用性，API 简洁，能自然融入 Go 项目。顶层包提供核心类型和构造函数；需要泛型自由函数时可按需导入对应子包。

## 特性

- **高性能**: 优化实现 — 双端队列两端操作摊销 O(1)、环形缓冲区底层，并审慎减少热点分配。
- **泛型支持**: 全面支持 Go 泛型，所有集合和工具包均提供类型安全的 API。
- **模块化设计**: 每个子包独立可用，也可通过统一入口访问。
- **简洁 API**: 一致的编码模式 — 方法链式调用、安全变体（`Try*`/`Peek*`）、符合习惯的错误处理。
- **测试完善**: 所有包均覆盖全面的单元测试。

## 包列表

### deque - 双端队列

基于环形缓冲区的高性能泛型双端队列。计入偶发扩缩容后，头尾操作均为摊销 O(1)。缓冲区在 push 时自动扩容，空闲时自动收缩。支持随机访问、旋转、任意位置插入和基于谓词的搜索。

```go
d := op.NewDeque[int](64)       // 预分配环形缓冲区容量
d.PushBack(1)                    // [1]
d.PushBack(2)                    // [1, 2]
d.PushFront(0)                   // [0, 1, 2]
d.PushBack(3)                    // [0, 1, 2, 3]

d.PopFront()                     // 返回 0，队空时 panic
d.Rotate(1)                      // 正向旋转：[2, 3, 1]
d.Insert(1, 99)                  // [2, 99, 3, 1]

// 安全访问，不会 panic
if v, ok := d.PeekFront(); ok {
    // 使用 v
}

// 按条件搜索
idx := d.Index(func(x int) bool { return x > 50 })  // idx=1 → 99
```

**文档**: [deque/README.md](deque/README.md) | [中文文档](deque/README_zh.md)

### emission - 事件发射器

带类型安全的事件发布订阅系统，支持异步即发即忘、并发发送并等待、以及完全同步的分发模式。通过订阅对象管理监听器生命周期，支持一次性监听器、panic 恢复和可配置的并发限制。

```go
// E = 事件类型（comparable），T = 负载类型
em := op.NewEmitter[string, int]()

// 订阅并管理生命周期
sub := em.On("order.created", func(amount int) {
    fmt.Printf("新订单：$%d\n", amount)
})
defer sub.Unsubscribe()

// 一次性监听器
em.Once("startup", func(v int) { fmt.Println("初始化完成") })

// 即发即忘（异步）
em.Emit("order.created", 150)

// 并发发送，等待所有监听器完成
em.EmitWait("order.created", 200)

// 同步发送，按注册顺序执行
em.EmitSync("order.created", 300)

// 从监听器 panic 中恢复
em.RecoverWith(func(event string, listener any, panicVal any) {
    log.Printf("监听器 panic，事件 %s: %v", event, panicVal)
})

// 限制并发监听器 goroutine 数量
em.SetConcurrency(4)
```

**文档**: [emission/README.md](emission/README.md)

### linq - LINQ 风格查询

受 .NET LINQ 启发的 Go 切片链式查询 API。提供过滤、投影、排序、分组、聚合、集合运算和连接操作，所有操作均即时求值。`Linq` 为值类型，多数链式方法返回新值。

```go
import (
    "github.com/wsshow/op"
    "github.com/wsshow/op/linq"
)

// --- 过滤与投影 ---
results := op.LinqFrom([]int{1, 2, 3, 4, 5, 6}).
    Where(func(x int) bool { return x%2 == 0 }).
    Select(func(x int) int { return x * 10 }).
    Results()
// results = [20, 40, 60]

// --- 多级排序 ---
users := []struct{ Name string; Age int }{{"Alice", 30}, {"Bob", 25}, {"Carol", 35}}
ordered := linq.OrderBy(op.LinqFrom(users),
    func(u struct{ Name string; Age int }) int { return u.Age },
).ThenByDescending(func(a, b struct{ Name string; Age int }) int {
    return strings.Compare(a.Name, b.Name)
})
for _, u := range ordered.Results() {
    fmt.Println(u.Name, u.Age)
}

// --- 聚合操作 ---
nums := op.LinqFrom([]int{10, 20, 30, 40})
sum := linq.Sum(nums)           // 100
avg := linq.Average(nums)       // 25.0
min, _ := linq.MinVal(nums)     // 10
cnt := nums.CountBy(func(x int) bool { return x > 20 })  // 2

// --- 集合运算（comparable 类型）---
a := op.LinqFrom([]int{1, 2, 3, 4})
b := op.LinqFrom([]int{3, 4, 5, 6})
union := linq.Union(a, b)        // [1, 2, 3, 4, 5, 6]
inter := linq.Intersect(a, b)    // [3, 4]
diff  := linq.Except(a, b)       // [1, 2]

// --- 分组 ---
words := op.LinqFrom([]string{"apple", "banana", "apricot", "blueberry", "avocado"})
groups := linq.GroupBy(words, func(w string) string { return string(w[0]) })
for _, g := range groups {
    fmt.Printf("Key %s: %v\n", g.Key, g.Items)
}

// --- 连接 ---
orders := op.LinqFrom([]struct{ ID, UserID int }{{1, 100}, {2, 200}})
customers := op.LinqFrom([]struct{ ID int; Name string }{{100, "Alice"}, {200, "Bob"}})
joined := linq.Join(
    orders, customers,
    func(o struct{ ID, UserID int }) int { return o.UserID },
    func(c struct{ ID int; Name string }) int { return c.ID },
    func(o struct{ ID, UserID int }, c struct{ ID int; Name string }) string {
        return fmt.Sprintf("订单 #%d 客户 %s", o.ID, c.Name)
    },
)
```

**文档**: [linq/README.md](linq/README.md)

### process - 进程管理

外部进程的生成、监控和生命周期管理工具。支持 stdout/stderr 逐行回调、带间隔门控的自动重启、基于 context 的取消和有限时停止等待，以及通过 `Manager` 实现的多进程编排。

```go
// --- 单个进程 ---
proc := op.NewProcess(op.Options{
    ExecPath: "my-server",
    Args:     []string{"--port", "8080", "--verbose"},
    Env:      []string{"LOG_LEVEL=debug"},
    OnStdout: func(line string) { log.Println("OUT:", line) },
    OnStderr: func(line string) { log.Println("ERR:", line) },
    OnBefore: func(p *op.Process) { log.Println("启动中...") },
    OnAfter:  func(p *op.Process) { log.Printf("已退出，退出码 %d", p.ExitCode()) },
})

if err := proc.Start(); err != nil {
    log.Fatal(err)
}

// 等待进程结束
<-proc.Done()
log.Printf("退出码: %d", proc.ExitCode())

// 带最小启动间隔控制的重启
proc.Restart()

// 发送信号
proc.Signal(os.Interrupt)

// 自定义停止超时
proc.StopWithTimeout(10 * time.Second)

// --- 多进程管理器 ---
mgr := op.NewProcessManager()

mgr.Add("api", op.Options{
    ExecPath: "./api-server",
    Args:     []string{"--port", "8080"},
})
mgr.Add("worker", op.Options{
    ExecPath: "./worker",
    Args:     []string{"--queue", "default"},
})

// 查询和控制
if p, ok := mgr.Get("api"); ok {
    log.Printf("API PID: %d", p.Pid())
}

// 遍历所有进程
mgr.Range(func(name string, p *op.Process) bool {
    log.Printf("%s: 运行中=%v", name, p.IsRunning())
    return true
})

// 批量操作
mgr.RestartAll()
defer mgr.StopAllWithTimeout(15 * time.Second)
```

**文档**: [process/README.md](process/README.md) | [中文文档](process/README_zh.md)

### slice - 切片工具

泛型切片包装器，提供函数式操作接口。支持 map、filter、reduce、任意位置插入/删除、排序、反转、合并和安全访问。多数修改方法返回 `*Slice` 以支持链式调用。

```go
import (
    "github.com/wsshow/op"
    "github.com/wsshow/op/slice"
)

s := op.NewSlice(1, 2, 3)

// --- 修改操作（原地修改，支持链式调用）---
s.Push(4, 5).Unshift(0)
// s = [0, 1, 2, 3, 4, 5]

val, ok := s.Pop()     // val=5, ok=true
val, ok = s.Shift()    // val=0, ok=true

s.Insert(2, 99)        // [1, 2, 99, 3, 4]

// --- 函数式操作（返回新 Slice）---
doubled := s.Map(func(x int) int { return x * 2 })
evens := s.Filter(func(x int) bool { return x%2 == 0 })

// --- 类型转换 ---
strs := slice.MapTo(s, func(x int) string { return strconv.Itoa(x) })
// strs 为 *Slice[string]

// --- 规约 ---
sum := s.Reduce(func(acc, cur int) int { return acc + cur }, 0)

// --- 排序 ---
s.Sort(func(a, b int) bool { return a < b })
s.Reverse()

// --- 安全访问 ---
if v, ok := s.Find(func(x int) bool { return x > 50 }); ok {
    // 使用 v
}
found := s.Some(func(x int) bool { return x > 3 })   // 任一匹配返回 true
allPos := s.Every(func(x int) bool { return x > 0 })  // 全部匹配返回 true

// --- 合并切片 ---
other := op.NewSlice(10, 20, 30)
merged := s.Concat(other)       // 新 Slice，原切片不变
```

**文档**: [slice/README.md](slice/README.md) | [中文文档](slice/README_zh.md)

### str - 字符串工具

字符串包装器，提供常用文本操作。多数方法原地修改并返回 `*String` 以支持链式调用。包含数值解析、Unicode 感知的反转、格式化以及 Python 式负索引的子串提取。

```go
s := op.NewString("  Hello, World!  ")

// --- 变换操作（原地修改，支持链式调用）---
s.TrimSpace().ToLower().ReplaceAll("world", "Gopher")
// s.String() = "hello, Gopher!"

// --- 检查操作 ---
s.Contains("Gopher")       // true
s.StartsWith("hello")      // true
s.Count("o")               // 2
s.Length()                 // 15
s.RuneLength()             // 15（Unicode 感知）

// --- 解析操作 ---
numStr := op.NewString("  42  ")
val, err := numStr.TrimSpace().ToInt()  // val=42

// --- 非修改操作（返回新 *String）---
cloned := s.Clone()
formatted := op.NewString("Hello, %s!").Format("World")  // "Hello, World!"
sub := s.Substring(7, 12)                                // "Gopher"
joined := op.JoinStrings([]string{"a", "b", "c"}, ",")   // "a,b,c"

// --- Unicode 感知操作 ---
op.NewString("こんにちは").Reverse()                     // "はちにんこ"
```

**文档**: [str/README.md](str/README.md)

### workerpool - 工作池

高性能 goroutine 池，限制并发数并将溢出任务排队。工作协程按需动态创建，空闲超时后自动回收。支持基于 context 的暂停/恢复、多种优雅关闭模式和 panic 恢复。

```go
wp := op.NewWorkerPool(4,  // 最大并发工作协程数
	 op.WithIdleTimeout(30*time.Second),
    op.WithPanicHandler(func(v any) {
        log.Printf("任务 panic: %v", v)
    }),
)

// 提交即发即忘的任务
for i := 0; i < 100; i++ {
    i := i
    wp.Submit(func() {
        // 处理任务 i
        time.Sleep(50 * time.Millisecond)
    })
}

// 提交任务并等待其完成
wp.SubmitWait(func() {
    // 关键的前置检查
    log.Println("前置检查完成")
})

// 查看队列压力
queued := wp.WaitingQueueSize()
log.Printf("等待中的任务数: %d", queued)

// 暂停所有工作协程
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
wp.Pause(ctx)   // 阻塞直至协程暂停或 context 到期

// 优雅关闭 — 完成所有排队任务后停止
wp.StopWait()

// 立即关闭 — 完成正在运行的任务，丢弃排队任务
// wp.Stop()
```

**文档**: [workerpool/README.md](workerpool/README.md) | [中文文档](workerpool/README_zh.md)

### generator - 生成器

基于 goroutine 和 channel 的轻量级协程式生成器。生成器函数通过 `Yield.Send()` 产出值。消费者用 `Next()` 获取值，并可选择在每次迭代时将结果回传给生成器，实现生产者与消费者之间的双向通信。

```go
// --- 基本值生成 ---
g := op.NewGenerator(func(yield op.Yield[int]) {
    for i := 0; i < 5; i++ {
        if yield.Stopped() {
            return  // 消费者请求停止
        }
        // yield.Send 阻塞直到消费者调用 Next
        result := yield.Send(i)
        fmt.Printf("生成器收到: %v\n", result)
    }
})

// 消费所有值
for {
    val, done := g.Next("ack")  // "ack" 回传给生成器
    if done {
        break
    }
    fmt.Printf("消费者获得: %d\n", val)
}

// --- 无限序列与提前停止 ---
fibGen := op.NewGenerator(func(yield op.Yield[int]) {
    a, b := 0, 1
    for {
        if yield.Stopped() {
            return
        }
        yield.Send(a)
        a, b = b, a+b
    }
})

// 获取前 10 个斐波那契数
for i := 0; i < 10; i++ {
    v, done := fibGen.Next()
    if done {
        break
    }
    fmt.Println(v)  // 0, 1, 1, 2, 3, 5, 8, 13, 21, 34
}
fibGen.Stop()
```

**文档**: [generator/README.md](generator/README.md) | [中文文档](generator/README_zh.md)

## 安装

```
go get github.com/wsshow/op
```

导入顶层包即可访问所有类型和构造函数：

```go
import "github.com/wsshow/op"
```

也可单独导入子包以减少依赖：

```go
import (
    "github.com/wsshow/op/deque"
    "github.com/wsshow/op/linq"
)
```

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/wsshow/op"
)

func main() {
    // 字符串链式操作
    s := op.NewString("  Hello, World!  ")
    s.TrimSpace().ToUpper().ReplaceAll("WORLD", "GOPHER")
    fmt.Println(s)  // "HELLO, GOPHER!"

    // 切片函数式操作
    sl := op.NewSlice(1, 2, 3, 4, 5).
        Filter(func(x int) bool { return x%2 != 0 }).
        Map(func(x int) int { return x * x })
    fmt.Println(sl.Data())  // [1, 9, 25]

    // 类型安全的事件发射器
    em := op.NewEmitter[string, string]()
    em.On("message", func(payload string) {
        fmt.Println("收到:", payload)
    })
    em.Emit("message", "来自发射器的问候")

    // LINQ 过滤与排序
    scores := op.LinqFrom([]int{85, 92, 78, 95, 88})
    passed := scores.
        Where(func(x int) bool { return x >= 80 }).
        Sort(func(a, b int) bool { return a < b })
    fmt.Println(passed.Results())                                     // [85, 88, 92, 95]
    fmt.Println("通过:", passed.Count(), "/", scores.Count())         // 通过: 4 / 5

    // 高性能双端队列
    d := op.NewDeque[string](8)
    d.PushBack("alpha")
    d.PushBack("beta")
    d.PushFront("omega")
    for d.Size() > 0 {
        fmt.Println(d.PopFront())
    }

    // 协程式生成器
    g := op.NewGenerator(func(yield op.Yield[int]) {
        for i := 1; i <= 3; i++ {
            yield.Send(i * 10)
        }
    })
    for {
        v, done := g.Next()
        if done {
            break
        }
        fmt.Println(v)  // 10, 20, 30
    }

    // 带 panic 恢复的工作池
    wp := op.NewWorkerPool(4, op.WithPanicHandler(func(v any) {
        log.Printf("从 panic 恢复: %v", v)
    }))
    for i := 0; i < 20; i++ {
        i := i
        wp.Submit(func() {
            fmt.Printf("任务 %d 执行中\n", i)
        })
    }
    wp.StopWait()

    // 进程管理
    proc := op.NewProcess(op.Options{
        ExecPath: "echo",
        Args:     []string{"hello"},
        OnStdout: func(line string) { fmt.Println("OUT:", line) },
    })
    if err := proc.Run(); err != nil {
        log.Fatal(err)
    }

    // 多进程管理器
    mgr := op.NewProcessManager()
    mgr.Add("healthcheck", op.Options{
        ExecPath: "curl",
        Args:     []string{"-s", "http://localhost:8080/health"},
    })
    defer mgr.StopAll()
}
```

## 目录结构

```
op/
├── deque/              # 泛型环形缓冲区双端队列
├── emission/           # 类型安全的事件发布/订阅
├── linq/               # LINQ 风格链式查询库
├── process/            # 外部进程生命周期管理
├── slice/              # 带函数式操作的泛型切片包装器
├── str/                # 支持链式调用的字符串包装器
├── workerpool/         # 有界 goroutine 池
├── generator/          # 协程式生成器
└── op.go               # 统一入口，含类型别名
```

## 贡献

欢迎贡献。请确保已有测试通过，并附上新功能的测试覆盖。提交较大的变更前，建议先提出 issue 讨论。

## 许可证

MIT - 详见 [LICENSE](LICENSE)。

## 致谢

- [deque](https://github.com/gammazero/deque) - 环形缓冲区双端队列的灵感来源
- [workerpool](https://github.com/gammazero/workerpool) - 工作池的灵感来源
- [emission](https://github.com/chuckpreslar/emission) - 事件发射器的灵感来源
