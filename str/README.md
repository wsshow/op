# Str - A String Wrapper with Utility Methods

English | [简体中文](README_zh.md)

`str` is a string wrapper in Go that provides a rich set of utility methods for common string operations, with a chainable API inspired by object-oriented programming patterns.

## Features

- **Chainable API**: Mutation methods return `*String` for method chaining
- **Rich Operations**: Contains, split, replace, trim, case conversion, and more
- **Type Conversion**: Convert to int, float with error handling and panic variants
- **Unicode-Aware**: Reverse, Substring, and RuneLength are rune-based
- **Immutable Derivation**: Substring, Concat, and Format return new instances without modifying the original
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
    s := str.New("Hello, World!")

    // Check if contains substring
    fmt.Println(s.Contains("World")) // Output: true

    // Find index
    fmt.Println(s.Index("World"))     // Output: 7
    fmt.Println(s.LastIndex("o"))     // Output: 8

    // Count occurrences
    fmt.Println(s.Count("o"))         // Output: 2
}
```

### String Manipulation with Chaining

```go
s := str.New("  Hello, World!  ")

// Chain multiple operations (mutation methods modify in-place)
result := s.TrimSpace().
    ReplaceAll("World", "Go").
    ToUpper()

fmt.Println(result.String()) // Output: HELLO, GO!
```

### Case Conversion

```go
s := str.New("Hello World")

fmt.Println(s.ToLower().String()) // Output: hello world
fmt.Println(s.ToUpper().String()) // Output: HELLO WORLD

// Case-insensitive comparison
s2 := str.New("Hello")
fmt.Println(s2.EqualFold("hello")) // Output: true
```

### Splitting

```go
s := str.New("apple,banana,cherry")

parts := s.Split(",")
fmt.Println(parts) // Output: [apple banana cherry]

// Split by whitespace
s2 := str.New("hello  world\tfoo")
fields := s2.Fields()
fmt.Println(fields) // Output: [hello world foo]
```

### Prefix and Suffix Checking

```go
s := str.New("https://example.com")

fmt.Println(s.StartsWith("https://")) // Output: true
fmt.Println(s.EndsWith(".com"))       // Output: true

// Trim prefix/suffix
s.TrimPrefix("https://")
fmt.Println(s.String()) // Output: example.com
```

### String Length

```go
s := str.New("Hello 世界")

// Byte length
fmt.Println(s.Length())      // Output: 12

// Unicode character count (rune length)
fmt.Println(s.RuneLength())  // Output: 8
```

### Trimming

```go
s := str.New("...Hello...")

// Trim specific characters
s.Trim(".")
fmt.Println(s.String()) // Output: Hello

// Trim whitespace
s2 := str.New("  spaces  ")
s2.TrimSpace()
fmt.Println(s2.String()) // Output: spaces

// Trim prefix/suffix
s3 := str.New("hello.txt")
s3.TrimSuffix(".txt")
fmt.Println(s3.String()) // Output: hello
```

### Concatenation

```go
s := str.New("Hello")

// Concat returns a NEW instance, original is unchanged
result := s.Concat(", ", "World", "!")
fmt.Println(result.String()) // Output: Hello, World!
fmt.Println(s.String())      // Output: Hello (unchanged)
```

### Reverse

```go
s := str.New("hello")
s.Reverse()
fmt.Println(s.String()) // Output: olleh

// Unicode-aware
s2 := str.New("你好世界")
s2.Reverse()
fmt.Println(s2.String()) // Output: 界世好你
```

### Substring

```go
s := str.New("Hello 世界")

// Extract substring (Unicode-aware, returns new instance)
sub := s.Substring(0, 7)
fmt.Println(sub.String()) // Output: Hello 世

// Original is unchanged
fmt.Println(s.String()) // Output: Hello 世界

// Negative end index (from end)
sub2 := str.New("hello.txt").Substring(0, -4)
fmt.Println(sub2.String()) // Output: hello
```

### Type Conversion

```go
// String to int
s1 := str.New("42")
num, err := s1.ToInt()
if err == nil {
    fmt.Println(num) // Output: 42
}

// MustInt panics on failure
s2 := str.New("42")
fmt.Println(s2.MustInt()) // Output: 42

// String to float
s3 := str.New("3.14")
f, err := s3.ToFloat()
if err == nil {
    fmt.Println(f) // Output: 3.14
}

// With automatic whitespace trimming
s4 := str.New("  123  ")
num, err = s4.ToInt()
fmt.Println(num) // Output: 123
```

### Formatting

```go
template := str.New("Hello, %s! You have %d messages.")

// Format returns a NEW instance, template is unchanged
formatted := template.Format("Alice", 5)

fmt.Println(formatted.String())
// Output: Hello, Alice! You have 5 messages.

fmt.Println(template.String())
// Output: Hello, %s! You have %d messages. (unchanged)
```

### Empty and Blank Checks

```go
s1 := str.New("")
s2 := str.New("text")
s3 := str.New("   ")

fmt.Println(s1.IsEmpty()) // Output: true
fmt.Println(s2.IsEmpty()) // Output: false
fmt.Println(s1.IsBlank()) // Output: true
fmt.Println(s3.IsBlank()) // Output: true
fmt.Println(s2.IsBlank()) // Output: false
```

### Cloning

```go
original := str.New("original")
clone := original.Clone()

clone.ToUpper()

fmt.Println(original.String()) // Output: original
fmt.Println(clone.String())    // Output: ORIGINAL
```

### Complex Example: Text Processing

```go
input := str.New("  HELLO@EXAMPLE.COM  ")

email := input.
    TrimSpace().
    ToLower()

if email.EndsWith("@example.com") && !email.IsEmpty() {
    username := email.Clone().
        ReplaceAll("@example.com", "").
        String()

    fmt.Printf("Username: %s\n", username) // Output: Username: hello
}
```

### Data Parsing

```go
data := str.New("Name:Alice,Age:25,City:NYC")

parts := data.Split(",")
for _, part := range parts {
    kv := str.New(part).Split(":")
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
- `New(s string) *String`: Create a new String instance

### Searching
- `Contains(substr string) bool`: Check if contains substring
- `Index(substr string) int`: Find first occurrence index (-1 if not found)
- `LastIndex(substr string) int`: Find last occurrence index (-1 if not found)
- `Count(substr string) int`: Count non-overlapping occurrences

### Splitting
- `Split(sep string) []string`: Split by separator
- `Fields() []string`: Split by whitespace

### Length
- `Length() int`: Get byte length
- `RuneLength() int`: Get Unicode character count (rune length)

### Modification (Chainable, mutate in-place)
- `ReplaceAll(old, new string) *String`: Replace all occurrences
- `Replace(old, new string, n int) *String`: Replace first n occurrences
- `Trim(cutset string) *String`: Trim characters from both ends
- `TrimSpace() *String`: Trim whitespace from both ends
- `TrimPrefix(prefix string) *String`: Remove prefix if present
- `TrimSuffix(suffix string) *String`: Remove suffix if present
- `ToLower() *String`: Convert to lowercase
- `ToUpper() *String`: Convert to uppercase
- `Reverse() *String`: Reverse the string (Unicode-aware)

### Derivation (Return new instance, original unchanged)
- `Concat(ss ...string) *String`: Concatenate strings
- `Substring(start, end int) *String`: Extract substring (Unicode-aware, negative end = from end)
- `Format(args ...interface{}) *String`: Format string with arguments

### Checking
- `StartsWith(prefix string) bool`: Check if starts with prefix
- `EndsWith(suffix string) bool`: Check if ends with suffix
- `EqualFold(t string) bool`: Case-insensitive equality
- `IsEmpty() bool`: Check if string is empty
- `IsBlank() bool`: Check if empty or only whitespace

### Conversion
- `ToInt() (int, error)`: Convert to integer
- `MustInt() int`: Convert to integer, panics on failure
- `ToFloat() (float64, error)`: Convert to float64
- `MustFloat() float64`: Convert to float64, panics on failure

### Utility
- `Clone() *String`: Create a copy
- `String() string`: Get underlying string value (satisfies fmt.Stringer)

## Design Notes

- **Mutation vs Derivation**: Methods that modify (Replace, Trim, ToLower, etc.) mutate in-place and return `*String` for chaining. Methods that derive (Concat, Substring, Format) return a **new** instance without affecting the original.
- **Unicode**: `Reverse()`, `Substring()`, and `RuneLength()` are Unicode-aware (rune-based).
- **Conversion**: `ToInt()` and `ToFloat()` automatically trim whitespace before parsing. Use `MustInt()`/`MustFloat()` for panic-on-failure variants.
- **Mutability**: Unlike Go's built-in immutable strings, mutation methods modify the internal state by design for chainability. Use `Clone()` if you need to preserve the original before mutation.

## License

MIT License
