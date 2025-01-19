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

type IgnorePattern struct {
	Pattern      string
	RegexPattern *regexp.Regexp
	ParentDirs   []string
	Negate       bool
}

// PatternMatcher provides methods to parse and match ignore patterns.
type PatternMatcher struct {
	ignorePatterns []IgnorePattern
}

// NewPatternMatcher initializes a new PatternMatcher instance from a list of patterns.
func NewPatternMatcher(patterns []string) (*PatternMatcher, error) {
	ignorePatterns, err := buildIgnorePatterns(patterns)
	if err != nil {
		return nil, err
	}
	return &PatternMatcher{
		ignorePatterns: ignorePatterns,
	}, nil
}

// NewPatternMatcherFromReader initializes a new PatternMatcher instance from an io.Reader.
func NewPatternMatcherFromReader(reader io.Reader) (*PatternMatcher, error) {
	patterns, err := internal.ReadLines(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse patterns: %w", err)
	}
	return NewPatternMatcher(patterns)
}

// NewPatternMatcherFromFile reads a file containing ignore patterns and returns a Parser instance.
func NewPatternMatcherFromFile(filepath string) (*PatternMatcher, error) {
	fileReader, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer fileReader.Close()

	patterns, err := internal.ReadLines(fileReader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse patterns: %w", err)
	}
	return NewPatternMatcher(patterns)
}

// Matches checks if the given file matches any pattern in the parser.
func (p *PatternMatcher) Matches(file string) (bool, error) {
	file = filepath.Clean(file)
	if file == "." {
		return false, nil
	}
	return matches(file, p.ignorePatterns)
}

func buildIgnorePatterns(patterns []string) ([]IgnorePattern, error) {
	var ignorePatterns []IgnorePattern

	for _, pattern := range patterns {
		trimmed := strings.TrimSpace(pattern)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		isNegation := strings.HasPrefix(trimmed, "!")
		if isNegation && len(trimmed) == 1 {
			// A single '!' is invalid
			return nil, errors.New("invalid pattern: '!'")
		}

		if isNegation {
			pattern = pattern[1:]
		}

		normalizedPattern := filepath.Clean(trimmed)
		patternDirs := strings.Split(pattern, string(filepath.Separator))

		regexPattern, err := internal.BuildRegex(normalizedPattern)
		if err != nil {
			return nil, err
		}

		ignorePatterns = append(ignorePatterns, IgnorePattern{
			Pattern:      normalizedPattern,
			RegexPattern: regexPattern,
			ParentDirs:   patternDirs,
			Negate:       isNegation,
		})
	}

	return ignorePatterns, nil
}

// matches checks if the file matches patterns efficiently.
func matches(file string, ignorePatterns []IgnorePattern) (bool, error) {
	// Normalize the file path to use OS-specific separators
	normalizedFile := filepath.FromSlash(file)

	// Split the parent path into components
	parentPath := filepath.Dir(normalizedFile)
	parentDirs := strings.Split(parentPath, string(filepath.Separator))

	matched := false

	for _, pattern := range ignorePatterns {
		matches, err := matchWithRegex(normalizedFile, pattern)
		if err != nil {
			return false, err
		}

		// If there's no direct match, check parent directories for a match
		if !matches && parentPath != "." && len(pattern.ParentDirs) <= len(parentDirs) {
			subPath := strings.Join(parentDirs[:len(pattern.ParentDirs)], string(filepath.Separator))
			subPathRegex, _ := internal.BuildRegex(subPath)
			subPathIgnorePattern := IgnorePattern{
				Pattern:      subPath,
				RegexPattern: subPathRegex,
			}
			matches, _ = matchWithRegex(strings.Join(pattern.ParentDirs, string(filepath.Separator)), subPathIgnorePattern)
		}

		// Update match status based on negation
		if matches {
			matched = !pattern.Negate
		}
	}

	return matched, nil
}

// matchWithRegex converts a pattern to a regular expression and checks if it matches the path.
func matchWithRegex(path string, ignorePattern IgnorePattern) (bool, error) {
	if _, err := filepath.Match(ignorePattern.Pattern, path); err != nil {
		return false, err
	}

	matched := ignorePattern.RegexPattern.MatchString(path)

	return matched, nil
}
