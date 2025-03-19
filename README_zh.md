# OP - Go 实用工具包集合

[English](./README.md) | 简体中文

`op` 是一个 Go 语言的实用工具集，它提供了多种可重用的包，用于处理常见的编程任务。每个包都被设计为轻量级、高效，并且易于集成到您的项目中。此仓库作为所有子包的集中入口点。

## 包含的包

该工具集包含以下包：

### deque

一个通用的双端队列（deque）实现。

- **特性**：支持从两端进行推入/弹出操作，兼容泛型。
- **使用方法**：详情请查看 [deque/README.md](deque/README.md) 或 [deque/README_zh.md](deque/README_zh.md)。

### emission

一个用于发布/订阅模式的通用事件发射器。

- **特性**：支持事件订阅，一次性监听器，异步/同步发射，以及恐慌恢复。
- **限制**：事件类型必须是可比较的。
- **使用方法**：详情请查看 [emission/README.md](emission/README.md)。

### linq

一个用于 Go 切片的 LINQ 风格查询库。

- **特性**：支持过滤、映射、排序、分组等多种操作。
- **使用方法**：详情请查看 [linq/README.md](linq/README.md)（如果存在）。

### process

用于管理外部进程的工具。

- **特性**：支持进程执行，标准输出/错误处理，以及进程管理。
- **文件**：
  - `process.go`：核心进程处理。
  - `process_m.go`：用于管理多个进程的进程管理器。
- **使用方法**：详情请查看 [process/README.md](process/README.md) 或 [process/README_zh.md](process/README_zh.md)。

### slice

一个带有实用方法的泛型切片包装器。

- **特性**：支持推入、弹出、过滤、映射、归约等操作。
- **使用方法**：详情请查看 [slice/README.md](slice/README.md) 或 [slice/README_zh.md](slice/README_zh.md)。

### str

一个带有常用操作的字符串包装器。

- **特性**：包含、拆分、替换、大小写转换等操作。
- **使用方法**：详情请查看 [str/README.md](str/README.md)（如果存在）。

### workerpool

一个用于并发任务执行的工人池。

- **特性**：固定大小的工人池，支持任务提交。
- **使用方法**：详情请查看 [workerpool/README.md](workerpool/README.md) 或 [workerpool/README_zh.md](workerpool/README_zh.md)。

### generator (未完成)

一个生成器包

- **文件**：出现在多个目录中，但缺乏清晰的实现。

## 安装

要在您的 Go 项目中使用`op`工具集，请运行以下命令：

```bash
go get github.com/wsshow/op
```

然后导入所需的包：

```go
import "github.com/wsshow/op"
```

## 使用示例

```go
package main
import (
	"fmt"
	"github.com/wsshow/op"
)
func main() {
	// 创建一个字符串
	s := op.NewString("Hello, World")
	fmt.Println(s.Contain("World")) // 输出：true
	// 创建一个切片
	sl := op.NewSlice(1, 2, 3)
	fmt.Println(sl.Data()) // 输出：[1 2 3]
	// 创建一个事件发射器
	em := op.NewEmitter[string]()
	em.On("event", func(args ...string) {
		fmt.Println("事件:", args)
	})
	em.Emit("event", "测试") // 输出：事件: [测试]
}
```

## 目录结构

```
op/
├── deque/              # 双端队列
├── emission/           # 事件发射器
├── linq/               # LINQ风格查询（多个实例）
├── process/            # 进程管理（多个实例）
├── slice/              # 切片工具
├── str/                # 字符串工具
├── workerpool/         # 工人池
└── op.go               # 工具集入口点
```
