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
	parentDirs   []string
	negate       bool
}

// PatternMatcher provides methods to parse, store, and evaluate ignore patterns against file paths.
type PatternMatcher struct {
	ignorePatterns []ignorePattern
}

// NewPatternMatcher initializes a new PatternMatcher instance from a list of string patterns.
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

// NewPatternMatcherFromFile reads a file containing ignore patterns and returns a PatternMatcher instance.
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

// Matches checks if the given file path matches any of the ignore patterns in the PatternMatcher.
func (p *PatternMatcher) Matches(file string) (bool, error) {
	file = filepath.Clean(file)
	if file == "." {
		return false, nil
	}
	return matches(file, p.ignorePatterns)
}

func buildIgnorePatterns(patterns []string) ([]ignorePattern, error) {
	var ignorePatterns []ignorePattern

	for _, pattern := range patterns {
		pattern := strings.TrimSpace(pattern)
		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}

		// normalize pattern
		pattern = filepath.Clean(pattern)
		isNegation := strings.HasPrefix(pattern, "!")
		if isNegation && len(pattern) == 1 {
			// A single '!' is invalid
			return nil, errors.New("invalid pattern: '!'")
		}

		if isNegation {
			pattern = pattern[1:]
		}

		patternDirs := strings.Split(pattern, "/")

		regexPattern, err := internal.BuildRegex(pattern)
		if err != nil {
			return nil, err
		}

		ignorePatterns = append(ignorePatterns, ignorePattern{
			pattern:      pattern,
			regexPattern: regexPattern,
			parentDirs:   patternDirs,
			negate:       isNegation,
		})
	}

	return ignorePatterns, nil
}

// matches checks if the file matches patterns efficiently.
func matches(file string, ignorePatterns []ignorePattern) (bool, error) {
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
		if !matches && parentPath != "." && len(pattern.parentDirs) <= len(parentDirs) {
			subPath := strings.Join(parentDirs[:len(pattern.parentDirs)], string(filepath.Separator))
			subPathRegex, _ := internal.BuildRegex(subPath)
			subPathIgnorePattern := ignorePattern{
				pattern:      subPath,
				regexPattern: subPathRegex,
			}
			matches, _ = matchWithRegex(strings.Join(pattern.parentDirs, string(filepath.Separator)), subPathIgnorePattern)
		}

		// Update match status based on negation
		if matches {
			matched = !pattern.negate
		}
	}

	return matched, nil
}

// matchWithRegex converts a pattern to a regular expression and checks if it matches the path.
func matchWithRegex(path string, ignorePat ignorePattern) (bool, error) {
	if _, err := filepath.Match(ignorePat.pattern, path); err != nil {
		return false, err
	}

	matched := ignorePat.regexPattern.MatchString(path)

	return matched, nil
}
