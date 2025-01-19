package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// ReadLines reads lines from an io.Reader and strips utf8 BOM characters..
func ReadLines(reader io.Reader) ([]string, error) {
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

func BuildRegex(pattern string) (*regexp.Regexp, error) {
	var regexBuilder strings.Builder
	regexBuilder.WriteString("^")

	// Traverse the pattern character by character and build equivalent regex
	for i := 0; i < len(pattern); i++ {
		char := pattern[i]

		switch char {
		case '*':
			// Check for "**" (double wildcard)
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				i++
				// Check for "**/" or trailing "**"
				if i+1 < len(pattern) && pattern[i+1] == '/' {
					i++
				}
				regexBuilder.WriteString("(.*)?") // Match zero or more directories
			} else {
				regexBuilder.WriteString("[^/]*") // Match zero or more characters except '/'
			}
		case '?':
			regexBuilder.WriteString("[^/]") // Match a single character except '/'
		case '.':
			regexBuilder.WriteString("\\.") // Escape '.' to match literal '.'
		case '$':
			regexBuilder.WriteString("\\$") // Escape '$' to match literal '$'
		case '\\':
			// Handle escaping
			if i+1 < len(pattern) {
				i++
				regexBuilder.WriteByte(pattern[i])
			} else {
				regexBuilder.WriteString("\\\\") // Escape the backslash itself
			}
		default:
			// Write other characters directly
			regexBuilder.WriteByte(char)
		}
	}

	// Complete the regex with an end anchor
	regexBuilder.WriteString("$")

	// Compile and match the regex
	return regexp.Compile(regexBuilder.String())
}
