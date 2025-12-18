# OP - Go 实用工具包集合

[![Go 版本](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![许可证](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

[English](./README.md) | 简体中文

`op` 是一个精心设计的 Go 工具包，提供了多种可重用的包来处理常见的编程任务。每个包都注重性能、易用性和泛型支持，可以轻松集成到您的项目中。本仓库作为所有子包的统一入口点。

## ✨ 特性

- 🚀 **高性能**: 优化的实现，注重内存和 CPU 效率
- 🎯 **泛型支持**: 全面支持 Go 1.18+ 泛型，提供类型安全的 API
- 📦 **模块化设计**: 每个包都是独立的，可按需使用
- 🔧 **易于集成**: 简洁的 API 设计，最小化学习成本
- 🧪 **完整测试**: 包含全面的单元测试

## 📦 包列表

### 🔄 deque - 双端队列

高性能的泛型双端队列实现，基于环形缓冲区设计。

- **特点**: O(1) 时间复杂度的两端操作、动态扩容、支持旋转、搜索等高级功能
- **适用场景**: 需要频繁在两端插入删除元素的场景
- **文档**: [deque/README.md](deque/README.md) | [中文文档](deque/README_zh.md)

### 📡 emission - 事件发射器

通用的事件发布订阅系统，支持异步和同步事件处理。

- **特点**: 基于唯一 ID 的监听器管理、支持一次性监听器、panic 恢复、事件类型必须是 `comparable`
- **适用场景**: 解耦组件间的通信、实现观察者模式
- **文档**: [emission/README.md](emission/README.md)

### 🔍 linq - LINQ 风格查询

为 Go 切片提供 LINQ 风格的链式查询 API。

- **特点**: 提供 30+ 种方法，包括 Where、Select、OrderBy、GroupBy、Distinct、First/Last、All/Any、Contains、Union、Intersect、Except、SelectMany、Chunk、TakeWhile、SkipWhile、Sum、Average 等
- **适用场景**: 复杂的数据转换和查询需求
- **文档**: [linq/README.md](linq/README.md)

### 🛠️ process - 进程管理

外部进程的创建、管理和执行工具。

- **特点**: 进程执行、标准输出/错误处理、多进程管理
- **核心文件**:
  - `process.go`: 核心进程处理
  - `process_m.go`: 多进程管理器
- **适用场景**: 需要执行和管理外部命令的场景
- **文档**: [process/README.md](process/README.md) | [中文文档](process/README_zh.md)

### 📋 slice - 切片工具

泛型切片包装器，提供丰富的实用方法。

- **特点**: Push、pop、filter、map、reduce、clear、clone 等操作
- **适用场景**: 增强切片的操作能力
- **文档**: [slice/README.md](slice/README.md) | [中文文档](slice/README_zh.md)

### 🔤 str - 字符串工具

字符串包装器，提供常见的字符串操作。

- **特点**: 包含检查、分割、替换、大小写转换等
- **适用场景**: 简化字符串处理逻辑
- **文档**: [str/README.md](str/README.md)

### ⚡ workerpool - 工作池

高性能的工作池实现，用于并发任务执行。

- **特点**: 动态工作线程管理、任务队列、暂停/恢复功能、自动资源回收
- **适用场景**: 控制并发度、提高任务处理效率
- **文档**: [workerpool/README.md](workerpool/README.md) | [中文文档](workerpool/README_zh.md)

### 🎲 generator - 生成器

轻量级的生成器实现，支持协程式的值生成。

- **特点**: 泛型支持、Yield 机制、安全的资源管理
- **适用场景**: 需要延迟计算或迭代生成值的场景
- **文档**: [generator/README.md](generator/README.md) | [中文文档](generator/README_zh.md)

## 🚀 安装

要在您的 Go 项目中使用 `op` 工具集，请运行以下命令：

```bash
go get github.com/wsshow/op
```

然后导入所需的包：

```go
import "github.com/wsshow/op"
```

## 💡 使用示例

```go
package main

import (
	"fmt"
	"github.com/wsshow/op"
)

func main() {
	// 创建字符串包装器
	s := op.NewString("Hello, World")
	fmt.Println(s.Contain("World")) // true

	// 创建泛型切片
	sl := op.NewSlice(1, 2, 3)
	fmt.Println(sl.Data()) // [1 2 3]

	// 创建事件发射器
	em := op.NewEmitter[string]()
	em.On("event", func(args ...string) {
		fmt.Println("事件:", args)
	})
	em.Emit("event", "测试") // 事件: [测试]

	// 创建双端队列
	d := op.NewDeque[int]()
	d.PushBack(1)
	d.PushFront(0)
	fmt.Println(d.PopFront()) // 0

	// 创建工作池
	wp := op.NewWorkerPool(4)
	wp.Submit(func() {
		fmt.Println("任务已执行")
	})
	wp.StopWait()
}
```

## 📁 目录结构

```
op/
├── deque/              # 双端队列实现
├── emission/           # 发布/订阅模式事件发射器
├── linq/               # LINQ 风格查询库
├── process/            # 进程管理工具
├── slice/              # 泛型切片工具
├── str/                # 字符串工具
├── workerpool/         # 并发工作池
├── generator/          # 生成器工具
└── op.go               # 主入口点
```

## 🤝 贡献

欢迎贡献代码！请随时提交 Pull Request。对于重大更改，请先创建 Issue 来讨论您想要更改的内容。

## 📄 许可证

本项目采用 MIT 许可证 - 详情请参阅 [LICENSE](LICENSE) 文件。

## 🙏 致谢

- [deque](https://github.com/gammazero/deque) - 双端队列实现的灵感来源
- [workerpool](https://github.com/gammazero/workerpool) - 工作池实现的灵感来源
- [emission](https://github.com/chuckpreslar/emission) - 事件发射器的灵感来源
