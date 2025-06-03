package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ReadLines reads lines from an io.Reader and strips UTF-8 BOM characters.
func ReadLines(reader io.Reader) ([]string, error) {
	if reader == nil {
		return nil, fmt.Errorf("reader cannot be nil")
	}

	scanner := bufio.NewScanner(reader)
	var lines []string
	utf8BOM := []byte{0xEF, 0xBB, 0xBF}

	for lineNumber := 0; scanner.Scan(); lineNumber++ {
		line := scanner.Bytes()
		if lineNumber == 0 {
			line = bytes.TrimPrefix(line, utf8BOM)
		}
		lines = append(lines, string(line))
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading lines: %w", err)
	}

	return lines, nil
}

// BuildRegex converts a gitignore-style pattern to a regular expression.
// It properly handles wildcards, escaping, and gitignore-specific rules.
func BuildRegex(pattern string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, fmt.Errorf("pattern cannot be empty")
	}

	var regexBuilder strings.Builder
	regexBuilder.WriteString("^")

	i := 0
	for i < len(pattern) {
		char := pattern[i]

		switch char {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				// Handle "**" double wildcard
				i++ // consume the second '*'

				// Check what follows the "**"
				if i+1 < len(pattern) && pattern[i+1] == '/' {
					// "**/" - matches zero or more directories
					i++ // consume the '/'
					regexBuilder.WriteString("(.*?/)?")
				} else if i+1 == len(pattern) {
					// "**" at end - matches anything
					regexBuilder.WriteString(".*")
				} else {
					// "**" followed by something else - treat as ".*"
					regexBuilder.WriteString(".*")
				}
			} else {
				// Single "*" - matches any characters except '/'
				regexBuilder.WriteString("[^/]*")
			}
		case '?':
			// Single character wildcard (except '/')
			regexBuilder.WriteString("[^/]")
		case '[':
			// Character class - find the closing bracket
			j := i + 1
			for j < len(pattern) && pattern[j] != ']' {
				j++
			}
			if j < len(pattern) {
				// Valid character class - write it as-is (it's already valid regex)
				charClass := pattern[i : j+1]
				regexBuilder.WriteString(charClass)
				i = j
			} else {
				// No closing bracket, treat as literal
				regexBuilder.WriteString("\\[")
			}
		case '.', '+', '^', '$', '(', ')', '{', '}', '|':
			// Escape regex metacharacters
			regexBuilder.WriteString("\\")
			regexBuilder.WriteByte(char)
		case '\\':
			// Handle escaping
			if i+1 < len(pattern) {
				i++
				nextChar := pattern[i]
				// Escape the next character
				if isRegexMetaChar(nextChar) {
					regexBuilder.WriteString("\\")
				}
				regexBuilder.WriteByte(nextChar)
			} else {
				// Trailing backslash
				regexBuilder.WriteString("\\\\")
			}
		default:
			// Regular character
			regexBuilder.WriteByte(char)
		}
		i++
	}

	regexBuilder.WriteString("$")

	regexStr := regexBuilder.String()
	regex, err := regexp.Compile(regexStr)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex %q: %w", regexStr, err)
	}

	return regex, nil
}

// isRegexMetaChar checks if a character has special meaning in regex
func isRegexMetaChar(char byte) bool {
	switch char {
	case '.', '+', '*', '?', '^', '$', '(', ')', '[', ']', '{', '}', '|', '\\':
		return true
	default:
		return false
	}
}

// NormalizePath normalizes a file path for consistent matching across platforms
func NormalizePath(path string) string {
	// Convert backslashes to forward slashes
	path = strings.ReplaceAll(path, "\\", "/")

	// Remove redundant slashes
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}

	// Remove leading "./"
	if strings.HasPrefix(path, "./") {
		path = path[2:]
	}

	// Remove trailing slash unless it's the root
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	return path
}
