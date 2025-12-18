# OP - A Collection of Go Utility Packages

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

English | [简体中文](README_zh.md)

`op` is a carefully crafted Go utility toolkit that provides a variety of reusable packages for common programming tasks. Each package focuses on performance, usability, and generic support, making them easy to integrate into your projects. This repository serves as a centralized entry point for all sub-packages.

## ✨ Features

- 🚀 **High Performance**: Optimized implementations with attention to memory and CPU efficiency
- 🎯 **Generic Support**: Full support for Go 1.18+ generics with type-safe APIs
- 📦 **Modular Design**: Each package is independent and can be used as needed
- 🔧 **Easy Integration**: Clean API design with minimal learning curve
- 🧪 **Fully Tested**: Comprehensive unit tests included

## 📦 Packages

### 🔄 deque - Double-Ended Queue

A high-performance generic double-ended queue implementation based on a circular buffer.

- **Features**: O(1) operations at both ends, dynamic resizing, rotation, search capabilities
- **Use Case**: Scenarios requiring frequent insertions/deletions at both ends
- **Docs**: [deque/README.md](deque/README.md) | [中文文档](deque/README_zh.md)

### 📡 emission - Event Emitter

A universal event publish-subscribe system supporting both async and sync event handling.

- **Features**: Unique ID-based listener management, one-time listeners, panic recovery, event type must be `comparable`
- **Use Case**: Decoupling component communication, implementing observer pattern
- **Docs**: [emission/README.md](emission/README.md)

### 🔍 linq - LINQ-Style Queries

LINQ-style chainable query API for Go slices.

- **Features**: 30+ methods including Where, Select, OrderBy, GroupBy, Distinct, First/Last, All/Any, Contains, Union, Intersect, Except, SelectMany, Chunk, TakeWhile, SkipWhile, Sum, Average, and more
- **Use Case**: Complex data transformation and query requirements
- **Docs**: [linq/README.md](linq/README.md)

### 🛠️ process - Process Management

Tools for creating, managing, and executing external processes.

- **Features**: Process execution, stdout/stderr handling, multi-process management
- **Core Files**:
  - `process.go`: Core process handling
  - `process_m.go`: Multi-process manager
- **Use Case**: Executing and managing external commands
- **Docs**: [process/README.md](process/README.md) | [中文文档](process/README_zh.md)

### 📋 slice - Slice Utilities

Generic slice wrapper with rich utility methods.

- **Features**: Push, pop, filter, map, reduce, clear, clone, and more operations
- **Use Case**: Enhancing slice manipulation capabilities
- **Docs**: [slice/README.md](slice/README.md) | [中文文档](slice/README_zh.md)

### 🔤 str - String Utilities

String wrapper with common string operations.

- **Features**: Contains check, split, replace, case conversion, etc.
- **Use Case**: Simplifying string processing logic
- **Docs**: [str/README.md](str/README.md)

### ⚡ workerpool - Worker Pool

High-performance worker pool for concurrent task execution.

- **Features**: Dynamic worker management, task queue, pause/resume, automatic resource cleanup
- **Use Case**: Controlling concurrency, improving task processing efficiency
- **Docs**: [workerpool/README.md](workerpool/README.md) | [中文文档](workerpool/README_zh.md)

### 🎲 generator - Generator

Lightweight generator implementation supporting coroutine-style value generation.

- **Features**: Generic support, yield mechanism, safe resource management
- **Use Case**: Lazy evaluation or iterative value generation
- **Docs**: [generator/README.md](generator/README.md) | [中文文档](generator/README_zh.md)

## 🚀 Installation

To use the `op` toolkit in your Go project, run:

```bash
go get github.com/wsshow/op
```

Then import the desired packages:

```go
import "github.com/wsshow/op"
```

## 💡 Usage Example

```go
package main

import (
	"fmt"
	"github.com/wsshow/op"
)

func main() {
	// Create a string wrapper
	s := op.NewString("Hello, World")
	fmt.Println(s.Contain("World")) // true

	// Create a generic slice
	sl := op.NewSlice(1, 2, 3)
	fmt.Println(sl.Data()) // [1 2 3]

	// Create an event emitter
	em := op.NewEmitter[string]()
	em.On("event", func(args ...string) {
		fmt.Println("Event:", args)
	})
	em.Emit("event", "test") // Event: [test]

	// Create a deque
	d := op.NewDeque[int]()
	d.PushBack(1)
	d.PushFront(0)
	fmt.Println(d.PopFront()) // 0

	// Create a worker pool
	wp := op.NewWorkerPool(4)
	wp.Submit(func() {
		fmt.Println("Task executed")
	})
	wp.StopWait()
}
```

## 📁 Directory Structure

```
op/
├── deque/              # Double-ended queue implementation
├── emission/           # Event emitter for pub/sub patterns
├── linq/               # LINQ-style query library
├── process/            # Process management tools
├── slice/              # Generic slice utilities
├── str/                # String utilities
├── workerpool/         # Concurrent worker pool
├── generator/          # Generator utilities
└── op.go               # Main entry point
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [deque](https://github.com/gammazero/deque) - Inspiration for the deque implementation
- [workerpool](https://github.com/gammazero/workerpool) - Inspiration for the worker pool implementation
- [emission](https://github.com/chuckpreslar/emission) - Inspiration for the event emitter
