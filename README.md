# go-dotignore

**go-dotignore** is a powerful Go library for parsing `.gitignore`-style files and matching file paths against specified ignore patterns. It provides full support for advanced ignore rules, negation patterns, and wildcards, making it an ideal choice for file exclusion in Go projects.

## Features

- Parse `.gitignore`-style files seamlessly
- Negation patterns (`!`) to override ignore rules
- Support for directories, files, and advanced wildcards like `**`
- Compatible with custom ignore files
- Does not process nested `.gitignore` files; all patterns are treated from a single source
- Fully compliant with the [`.gitignore` specification](https://git-scm.com/docs/gitignore)
- Lightweight API built on Go best practices

## Installation

To install **go-dotignore** in your Go project, run:

```bash
go get github.com/codeglyph/go-dotignore
```

## Getting Started

### Example: Basic Usage

Here is a simple example of how to use **go-dotignore**:

```go
package main

import (
 "fmt"
 "log"
 "github.com/codeglyph/go-dotignore"
)

func main() {
 // Define ignore patterns
 patterns := []string{
  "*.log",
  "!important.log",
  "temp/",
 }

 // Create a new pattern matcher
 matcher, err := dotignore.NewPatternMatcher(patterns)
 if err != nil {
  log.Fatalf("Failed to create pattern matcher: %v", err)
 }

 // Check if a file matches the patterns
 isIgnored, err := matcher.Matches("debug.log")
 if err != nil {
  log.Fatalf("Error matching file: %v", err)
 }
 fmt.Printf("Should ignore 'debug.log': %v\n", isIgnored)

 isIgnored, err = matcher.Matches("important.log")
 if err != nil {
  log.Fatalf("Error matching file: %v", err)
 }
 fmt.Printf("Should ignore 'important.log': %v\n", isIgnored)
}
```

### Example: Parsing a File

To parse patterns from a file, use the `NewPatternMatcherFromFile` method:

```go
package main

import (
 "log"
 "github.com/codeglyph/go-dotignore"
)

func main() {
 matcher, err := dotignore.NewPatternMatcherFromFile(".ignore")
 if err != nil {
  log.Fatalf("Failed to parse ignore file: %v", err)
 }

 isIgnored, err := matcher.Matches("example.txt")
 if err != nil {
  log.Fatalf("Error matching file: %v", err)
 }
 log.Printf("Should ignore 'example.txt': %v", isIgnored)
}
```

### Example: Parsing from Reader

To parse patterns from an `io.Reader`, use the `NewPatternMatcherFromReader` method:

```go
package main

import (
 "bytes"
 "log"
 "github.com/codeglyph/go-dotignore"
)

func main() {
 reader := bytes.NewBufferString("**/temp\n!keep/")
 matcher, err := dotignore.NewPatternMatcherFromReader(reader)
 if err != nil {
  log.Fatalf("Failed to parse patterns from reader: %v", err)
 }

 isIgnored, err := matcher.Matches("temp/file.txt")
 if err != nil {
  log.Fatalf("Error matching file: %v", err)
 }
 log.Printf("Should ignore 'temp/file.txt': %v", isIgnored)
}
```

## Advanced Features

### Negation Patterns

Negation patterns (`!`) allow you to override ignore rules. For example:

- `*.log` ignores all `.log` files.
- `!important.log` includes `important.log` even though `.log` files are ignored.

### Wildcard Support

- `*` matches any string except `/`.
- `?` matches any single character except `/`.
- `**` matches any number of directories.

### Directory Matching

- `dir/` matches only directories named `dir`.
- `dir/**` matches everything inside `dir` recursively.

### Custom Ignore Files

**go-dotignore** supports custom ignore files. Simply provide the file path or patterns programmatically using `NewPatternMatcherFromFile` or `NewPatternMatcher`.

### Non-Nested Processing

The library does not automatically process nested `.gitignore` files or directory-level `.ignore` files. All patterns are treated as coming from a single source file or list.

### Specification Compliance

**go-dotignore** follows the [`.gitignore` specification](https://git-scm.com/docs/gitignore) closely, ensuring consistent behavior with Git's pattern matching rules.

## Contributing

We welcome contributions to **go-dotignore**! Here's how you can contribute:

1. Fork the repository.
2. Create a new branch for your changes.
3. Add tests for your changes (if applicable).
4. Run all tests to ensure nothing breaks.
5. Submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for more details.

## Acknowledgements

This library is inspired by Git's `.gitignore` pattern matching and aims to bring the same functionality to Go projects.
