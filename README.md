# OP - A Collection of Go Utility Packages

English | [简体中文](README_zh.md)

`op` is a Go utility toolkit that provides a variety of reusable packages for common programming tasks. Each package is designed to be lightweight, efficient, and easy to integrate into your projects. This repository serves as a centralized entry point for all sub-packages.

## Packages

The toolkit includes the following packages:

### deque

A generic double-ended queue (deque) implementation.

- **Features**: Push/pop from both ends, generic support.
- **Usage**: See [deque/README.md](deque/README.md) or [deque/README_zh.md](deque/README_zh.md) for details.

### emission

A generic event emitter for pub/sub patterns.

- **Features**: Event subscription, once listeners, async/sync emission, panic recovery.
- **Constraints**: Event type must be `comparable`.
- **Usage**: See [emission/README.md](emission/README.md).

### linq

A LINQ-style query library for Go slices.

- **Features**: Filtering, mapping, sorting, grouping, and more.
- **Usage**: See [linq/README.md](linq/README.md) (if exists).

### process

Tools for managing external processes.

- **Features**: Process execution, stdout/stderr handling, process management.
- **Files**:
  - `process.go`: Core process handling.
  - `process_m.go`: Process manager for multiple processes.
- **Usage**: See [process/README.md](process/README.md) or [process/README_zh.md](process/README_zh.md).

### slice

A generic slice wrapper with utility methods.

- **Features**: Push, pop, filter, map, reduce, etc.
- **Usage**: See [slice/README.md](slice/README.md) or [slice/README_zh.md](slice/README_zh.md).

### str

A string wrapper with common operations.

- **Features**: Contains, split, replace, case conversion, etc.
- **Usage**: See [str/README.md](str/README.md) (if exists).

### workerpool

A worker pool for concurrent task execution.

- **Features**: Fixed-size worker pool, task submission.
- **Usage**: See [workerpool/README.md](workerpool/README.md) or [workerpool/README_zh.md](workerpool/README_zh.md).

### generator (Incomplete)

A generator package

## Installation

To use the `op` toolkit in your Go project, run:

```bash
go get github.com/wsshow/op
```

Then import the desired packages:

```go
import "github.com/wsshow/op"
```

## Usage Example

```go
package main

import (
	"fmt"
	"github.com/wsshow/op"
)

func main() {
	// Create a string
	s := op.NewString("Hello, World")
	fmt.Println(s.Contain("World")) // true

	// Create a slice
	sl := op.NewSlice(1, 2, 3)
	fmt.Println(sl.Data()) // [1 2 3]

	// Create an event emitter
	em := op.NewEmitter[string]()
	em.On("event", func(args ...string) {
		fmt.Println("Event:", args)
	})
	em.Emit("event", "test") // Event: [test]
}
```

## Directory Structure

```
op/
├── deque/              # Double-ended queue
├── emission/           # Event emitter
├── linq/               # LINQ-style queries (multiple instances)
├── process/            # Process management (multiple instances)
├── slice/              # Slice utilities (multiple instances)
├── str/                # String utilities
├── workerpool/         # Worker pool
└── op.go               # Toolkit entry point
```
