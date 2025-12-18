# Str - A String Wrapper with Utility Methods

English | [简体中文](README_zh.md)

`str` is a string wrapper in Go that provides a rich set of utility methods for common string operations, with a chainable API inspired by object-oriented programming patterns.

## Features

- **Chainable API**: Most methods return `*String` for method chaining
- **Rich Operations**: Contains, split, replace, trim, case conversion, and more
- **Type Conversion**: Convert to int, float with built-in error handling
- **Unicode-Aware**: Separate byte and rune length methods
- **Familiar Methods**: API inspired by common string operations in other languages

## Installation

```bash
go get github.com/wsshow/op/str
```

## Usage Examples

### Basic String Operations

```go
package main

import (
    "fmt"
    "github.com/wsshow/op/str"
)

func main() {
    s := str.NewString("Hello, World!")
    
    // Check if contains substring
    fmt.Println(s.Contain("World")) // Output: true
    
    // Find index
    fmt.Println(s.Index("World"))     // Output: 7
    fmt.Println(s.LastIndex("o"))     // Output: 8
}
```

### String Manipulation with Chaining

```go
s := str.NewString("  Hello, World!  ")

// Chain multiple operations
result := s.TrimSpace().
    ReplaceAll("World", "Go").
    ToUpper()

fmt.Println(result.String()) // Output: HELLO, GO!
```

### Case Conversion

```go
s := str.NewString("Hello World")

fmt.Println(s.ToLower().String()) // Output: hello world
fmt.Println(s.ToUpper().String()) // Output: HELLO WORLD
```

### Splitting

```go
s := str.NewString("apple,banana,cherry")

parts := s.Split(",")
fmt.Println(parts) // Output: [apple banana cherry]
```

### Prefix and Suffix Checking

```go
s := str.NewString("hello.txt")

fmt.Println(s.StartsWith("hello")) // Output: true
fmt.Println(s.EndsWith(".txt"))     // Output: true
fmt.Println(s.EndsWith(".pdf"))     // Output: false
```

### String Length

```go
s := str.NewString("Hello 世界")

// Byte length
fmt.Println(s.Length())      // Output: 12

// Unicode character count (rune length)
fmt.Println(s.RuneLength())  // Output: 8
```

### Trimming

```go
s := str.NewString("...Hello...")

// Trim specific characters
s.Trim(".")
fmt.Println(s.String()) // Output: Hello

// Trim whitespace
s2 := str.NewString("  spaces  ")
s2.TrimSpace()
fmt.Println(s2.String()) // Output: spaces
```

### Concatenation

```go
s := str.NewString("Hello")

s.Concat(", ", "World", "!")
fmt.Println(s.String()) // Output: Hello, World!
```

### Substring

```go
s := str.NewString("Hello 世界")

// Extract substring (Unicode-aware)
sub := s.Substring(0, 7)
fmt.Println(sub.String()) // Output: Hello 世
```

### Type Conversion

```go
// String to int
s1 := str.NewString("42")
num, err := s1.ToInt()
if err == nil {
    fmt.Println(num) // Output: 42
}

// String to float
s2 := str.NewString("3.14")
f, err := s2.ToFloat()
if err == nil {
    fmt.Println(f) // Output: 3.14
}

// With whitespace trimming
s3 := str.NewString("  123  ")
num, err = s3.ToInt() // Automatically trims whitespace
fmt.Println(num)       // Output: 123
```

### Formatting

```go
template := str.NewString("Hello, %s! You have %d messages.")
formatted := template.Format("Alice", 5)

fmt.Println(formatted.String()) 
// Output: Hello, Alice! You have 5 messages.
```

### Checking Empty String

```go
s1 := str.NewString("")
s2 := str.NewString("text")

fmt.Println(s1.IsEmpty()) // Output: true
fmt.Println(s2.IsEmpty()) // Output: false
```

### Cloning

```go
original := str.NewString("original")
clone := original.Clone()

clone.ToUpper()

fmt.Println(original.String()) // Output: original
fmt.Println(clone.String())    // Output: ORIGINAL
```

### Complex Example: Text Processing

```go
// Process user input
input := str.NewString("  HELLO@EXAMPLE.COM  ")

email := input.
    TrimSpace().
    ToLower().
    Clone()

if email.EndsWith("@example.com") && !email.IsEmpty() {
    username := email.
        ReplaceAll("@example.com", "").
        String()
    
    fmt.Printf("Username: %s\n", username) // Output: Username: hello
}
```

### URL Processing

```go
url := str.NewString("https://example.com/api/v1/users")

if url.StartsWith("https://") {
    path := str.NewString(url.String())
    path.ReplaceAll("https://example.com", "")
    fmt.Println(path.String()) // Output: /api/v1/users
}
```

### Data Parsing

```go
data := str.NewString("Name:Alice,Age:25,City:NYC")

parts := data.Split(",")
for _, part := range parts {
    kv := str.NewString(part).Split(":")
    if len(kv) == 2 {
        fmt.Printf("%s = %s\n", kv[0], kv[1])
    }
}
// Output:
// Name = Alice
// Age = 25
// City = NYC
```

## API Overview

### Creation
- `NewString(s string) *String`: Create a new String instance

### Searching
- `Contain(substr string) bool`: Check if contains substring
- `Index(substr string) int`: Find first occurrence index (-1 if not found)
- `LastIndex(substr string) int`: Find last occurrence index (-1 if not found)

### Splitting
- `Split(sep string) []string`: Split into slice

### Length
- `Length() int`: Get byte length
- `RuneLength() int`: Get Unicode character count (rune length)

### Modification (Chainable)
- `ReplaceAll(old, new string) *String`: Replace all occurrences
- `Trim(cutset string) *String`: Trim characters from both ends
- `TrimSpace() *String`: Trim whitespace from both ends
- `ToLower() *String`: Convert to lowercase
- `ToUpper() *String`: Convert to uppercase
- `Concat(ss ...string) *String`: Concatenate strings
- `Substring(start, end int) *String`: Extract substring (Unicode-aware)

### Checking
- `StartsWith(prefix string) bool`: Check if starts with prefix
- `EndsWith(suffix string) bool`: Check if ends with suffix
- `IsEmpty() bool`: Check if string is empty

### Conversion
- `ToInt() (int, error)`: Convert to integer
- `ToFloat() (float64, error)`: Convert to float64
- `Format(args ...interface{}) *String`: Format string with arguments

### Utility
- `Clone() *String`: Create a copy
- `String() string`: Get underlying string value

## Notes

- **Chaining**: Methods that modify the string return `*String` for chaining
- **Mutability**: Unlike Go's built-in strings, String methods modify the internal state
- **Unicode**: `Substring()` and `RuneLength()` are Unicode-aware (rune-based)
- **Conversion**: `ToInt()` and `ToFloat()` automatically trim whitespace before parsing

## License

MIT License
