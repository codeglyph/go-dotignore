package dotignore

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/codeglyph/go-dotignore/internal"
)

type ignorePattern struct {
	pattern      string
	regexPattern *regexp.Regexp
	isDirectory  bool // true if pattern ends with /
	negate       bool
	hasWildcard  bool // true if pattern contains wildcards
}

// PatternMatcher provides methods to parse, store, and evaluate ignore patterns against file paths.
type PatternMatcher struct {
	ignorePatterns []ignorePattern
}

// NewPatternMatcher initializes a new PatternMatcher instance from a list of string patterns.
func NewPatternMatcher(patterns []string) (*PatternMatcher, error) {
	ignorePatterns, err := buildIgnorePatterns(patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to build ignore patterns: %w", err)
	}
	return &PatternMatcher{
		ignorePatterns: ignorePatterns,
	}, nil
}

// NewPatternMatcherFromReader initializes a new PatternMatcher instance from an io.Reader.
func NewPatternMatcherFromReader(reader io.Reader) (*PatternMatcher, error) {
	if reader == nil {
		return nil, errors.New("reader cannot be nil")
	}

	patterns, err := internal.ReadLines(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse patterns from reader: %w", err)
	}
	return NewPatternMatcher(patterns)
}

// NewPatternMatcherFromFile reads a file containing ignore patterns and returns a PatternMatcher instance.
func NewPatternMatcherFromFile(filePath string) (*PatternMatcher, error) {
	if filePath == "" {
		return nil, errors.New("file path cannot be empty")
	}

	fileReader, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %q: %w", filePath, err)
	}
	defer fileReader.Close()

	patterns, err := internal.ReadLines(fileReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse patterns from file %q: %w", filePath, err)
	}
	return NewPatternMatcher(patterns)
}

// Matches checks if the given file path matches any of the ignore patterns in the PatternMatcher.
// It returns true if the file should be ignored, false otherwise.
func (p *PatternMatcher) Matches(file string) (bool, error) {
	if file == "" {
		return false, nil
	}

	// Clean and normalize the path
	file = filepath.Clean(file)
	if file == "." || file == "./" {
		return false, nil
	}

	// Convert backslashes to forward slashes for consistent matching
	// Use explicit conversion to handle all cases
	file = strings.ReplaceAll(file, "\\", "/")

	return p.matchesInternal(file)
}

func buildIgnorePatterns(patterns []string) ([]ignorePattern, error) {
	var ignorePatterns []ignorePattern

	for i, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)

		// Skip empty lines and comments
		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}

		// Handle negation
		isNegation := strings.HasPrefix(pattern, "!")
		if isNegation {
			if len(pattern) == 1 {
				return nil, fmt.Errorf("invalid pattern at line %d: single '!' is not allowed", i+1)
			}
			pattern = pattern[1:]
		}

		// Convert backslashes to forward slashes for consistent handling
		// filepath.ToSlash might not handle all cases, so we'll be explicit
		pattern = strings.ReplaceAll(pattern, "\\", "/")

		// Check if pattern is for directories only (after normalization)
		isDirectory := strings.HasSuffix(pattern, "/")
		if isDirectory {
			pattern = strings.TrimSuffix(pattern, "/")
		}

		// Validate pattern is not empty after processing
		if pattern == "" {
			return nil, fmt.Errorf("invalid pattern at line %d: pattern cannot be empty", i+1)
		}

		// Check if pattern contains wildcards
		hasWildcard := strings.ContainsAny(pattern, "*?")

		// Build regex pattern
		regexPattern, err := internal.BuildRegex(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to build regex for pattern %q at line %d: %w", pattern, i+1, err)
		}

		ignorePatterns = append(ignorePatterns, ignorePattern{
			pattern:      pattern,
			regexPattern: regexPattern,
			isDirectory:  isDirectory,
			negate:       isNegation,
			hasWildcard:  hasWildcard,
		})
	}

	return ignorePatterns, nil
}

// matchesInternal performs the actual pattern matching logic
func (p *PatternMatcher) matchesInternal(file string) (bool, error) {
	matched := false

	for _, pattern := range p.ignorePatterns {
		isMatch, err := p.matchPattern(file, pattern)
		if err != nil {
			return false, fmt.Errorf("error matching pattern %q against file %q: %w", pattern.pattern, file, err)
		}

		if isMatch {
			matched = !pattern.negate
		}
	}

	return matched, nil
}

// matchPattern checks if a file matches a specific pattern
func (p *PatternMatcher) matchPattern(file string, pattern ignorePattern) (bool, error) {
	// Try the regex pattern first
	if pattern.regexPattern.MatchString(file) {
		return true, nil
	}

	// Special handling for directory patterns
	if pattern.isDirectory {
		// Pattern like "build/" should match "build/" and anything inside "build/"
		dirName := pattern.pattern
		if file == dirName+"/" || file == dirName {
			return true, nil
		}
		if strings.HasPrefix(file, dirName+"/") {
			return true, nil
		}
	}

	// For patterns with wildcards, try matching parts of the path
	if pattern.hasWildcard {
		parts := strings.Split(file, "/")

		// For patterns like "src/*.txt", try matching against subpaths
		for i := 0; i < len(parts); i++ {
			subPath := strings.Join(parts[i:], "/")
			if pattern.regexPattern.MatchString(subPath) {
				return true, nil
			}
		}

		// Also try matching the full path from different starting points
		for i := 0; i < len(parts); i++ {
			if i == 0 {
				continue // already tried full path above
			}
			prefixPath := strings.Join(parts[:i], "/")
			remainingPath := strings.Join(parts[i:], "/")

			// Check if pattern could match from this point
			if pattern.regexPattern.MatchString(prefixPath + "/" + remainingPath) {
				return true, nil
			}
		}
	}

	// For patterns with path separators, try matching as substring
	if strings.Contains(pattern.pattern, "/") {
		// Pattern like "src/test.txt" should match exactly or as part of path
		if file == pattern.pattern {
			return true, nil
		}
		if strings.Contains(file, pattern.pattern) {
			return true, nil
		}

		// Try matching with different path boundaries
		if strings.HasSuffix(file, "/"+pattern.pattern) || strings.HasSuffix(file, pattern.pattern) {
			return true, nil
		}
	}

	// For simple patterns (no path separators), check filename components
	if !strings.Contains(pattern.pattern, "/") {
		parts := strings.Split(file, "/")
		for _, part := range parts {
			if pattern.regexPattern.MatchString(part) {
				return true, nil
			}
		}
	}

	return false, nil
}
