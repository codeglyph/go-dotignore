package dotignore_test

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/codeglyph/go-dotignore"
)

func ExampleNewPatternMatcher() {
	patterns := []string{"*.log", "!important.log", "temp/"}
	matcher, err := dotignore.NewPatternMatcher(patterns)
	if err != nil {
		log.Fatalf("Failed to create pattern matcher: %v", err)
	}

	file := "debug.log"
	matches, err := matcher.Matches(file)
	if err != nil {
		log.Fatalf("Error matching file: %v", err)
	}

	fmt.Printf("%s matches: %v\n", file, matches)

	importantFile := "important.log"
	matches, err = matcher.Matches(importantFile)
	if err != nil {
		log.Fatalf("Error matching file: %v", err)
	}

	fmt.Printf("%s matches: %v\n", importantFile, matches)
	// Output:
	// debug.log matches: true
	// important.log matches: false
}

func ExamplePatternMatcher_Matches() {
	patterns := []string{"*.txt", "reports/"}
	matcher, err := dotignore.NewPatternMatcher(patterns)
	if err != nil {
		log.Fatalf("Failed to create pattern matcher: %v", err)
	}

	files := []string{"notes.txt", "data.json", "reports/summary.pdf", "images/picture.jpg"}

	for _, file := range files {
		matches, err := matcher.Matches(file)
		if err != nil {
			log.Printf("Error matching file %s: %v", file, err)
			continue
		}

		fmt.Printf("%s matches: %v\n", file, matches)
	}
	// Output:
	// notes.txt matches: true
	// data.json matches: false
	// reports/summary.pdf matches: true
	// images/picture.jpg matches: false
}

func ExampleNewPatternMatcherFromReader() {
	reader := strings.NewReader("*.log\n!important.log\ntemp/")
	matcher, err := dotignore.NewPatternMatcherFromReader(reader)
	if err != nil {
		log.Fatalf("Failed to create pattern matcher: %v", err)
	}

	file := "debug.log"
	matches, err := matcher.Matches(file)
	if err != nil {
		log.Fatalf("Error matching file: %v", err)
	}

	fmt.Printf("%s matches: %v\n", file, matches)

	importantFile := "important.log"
	matches, err = matcher.Matches(importantFile)
	if err != nil {
		log.Fatalf("Error matching file: %v", err)
	}

	fmt.Printf("%s matches: %v\n", importantFile, matches)
	// Output:
	// debug.log matches: true
	// important.log matches: false
}

func ExampleNewPatternMatcherFromFile() {
	// Create a temporary file to simulate the test.gitignore file
	fileContent := "*.log\n!important.log\ntemp/"
	fileName := "test.gitignore"
	err := os.WriteFile(fileName, []byte(fileContent), 0644)
	if err != nil {
		log.Fatalf("Failed to create test.gitignore file: %v", err)
	}
	defer os.Remove(fileName) // Ensure the file is cleaned up after the test

	matcher, err := dotignore.NewPatternMatcherFromFile(fileName)
	if err != nil {
		log.Fatalf("Failed to create pattern matcher from file: %v", err)
	}

	file := "debug.log"
	matches, err := matcher.Matches(file)
	if err != nil {
		log.Fatalf("Error matching file: %v", err)
	}

	fmt.Printf("%s matches: %v\n", file, matches)

	importantFile := "important.log"
	matches, err = matcher.Matches(importantFile)
	if err != nil {
		log.Fatalf("Error matching file: %v", err)
	}

	fmt.Printf("%s matches: %v\n", importantFile, matches)
	// Output:
	// debug.log matches: true
	// important.log matches: false
}

// ExamplePatternMatcher_Matches_directories demonstrates directory pattern matching
func ExamplePatternMatcher_Matches_directories() {
	patterns := []string{"build/", "*.tmp", "logs/**"}
	matcher, err := dotignore.NewPatternMatcher(patterns)
	if err != nil {
		log.Fatalf("Failed to create pattern matcher: %v", err)
	}

	files := []string{
		"build/",               // Directory
		"build/app.js",         // File in ignored directory
		"cache.tmp",            // Temporary file
		"logs/app.log",         // File in logs directory
		"logs/debug/error.log", // Nested file in logs
		"src/main.go",          // Regular source file
	}

	for _, file := range files {
		matches, err := matcher.Matches(file)
		if err != nil {
			log.Printf("Error matching file %s: %v", file, err)
			continue
		}
		fmt.Printf("%-20s matches: %v\n", file, matches)
	}
	// Output:
	// build/               matches: true
	// build/app.js         matches: true
	// cache.tmp            matches: true
	// logs/app.log         matches: true
	// logs/debug/error.log matches: true
	// src/main.go          matches: false
}

// ExamplePatternMatcher_Matches_wildcards demonstrates wildcard pattern matching
func ExamplePatternMatcher_Matches_wildcards() {
	patterns := []string{
		"**/*.test.js",   // Test files anywhere
		"src/*/index.js", // Index files in immediate subdirs of src
		"file?.txt",      // Single character wildcard
	}
	matcher, err := dotignore.NewPatternMatcher(patterns)
	if err != nil {
		log.Fatalf("Failed to create pattern matcher: %v", err)
	}

	files := []string{
		"app.test.js",              // Root level test
		"src/utils/helper.test.js", // Nested test file
		"src/components/index.js",  // Index in component dir
		"src/utils/other.js",       // Non-index file in utils
		"file1.txt",                // Single char wildcard match
		"file10.txt",               // Multiple chars - no match
	}

	for _, file := range files {
		matches, err := matcher.Matches(file)
		if err != nil {
			log.Printf("Error matching file %s: %v", file, err)
			continue
		}
		fmt.Printf("%-25s matches: %v\n", file, matches)
	}
	// Output:
	// app.test.js               matches: true
	// src/utils/helper.test.js  matches: true
	// src/components/index.js   matches: true
	// src/utils/other.js        matches: false
	// file1.txt                 matches: true
	// file10.txt                matches: false
}

// ExamplePatternMatcher_Matches_negation demonstrates negation pattern behavior
func ExamplePatternMatcher_Matches_negation() {
	patterns := []string{
		"*.log",            // Ignore all log files
		"!important.log",   // But keep important.log
		"build/**",         // Ignore everything in build
		"!build/README.md", // But keep the README
		"!build/docs/**",   // And keep all docs
	}
	matcher, err := dotignore.NewPatternMatcher(patterns)
	if err != nil {
		log.Fatalf("Failed to create pattern matcher: %v", err)
	}

	files := []string{
		"app.log",              // Regular log file
		"important.log",        // Negated log file
		"build/app.js",         // Build artifact
		"build/README.md",      // Negated build file
		"build/docs/api.md",    // Negated docs file
		"build/dist/bundle.js", // Still ignored build file
	}

	for _, file := range files {
		matches, err := matcher.Matches(file)
		if err != nil {
			log.Printf("Error matching file %s: %v", file, err)
			continue
		}
		fmt.Printf("%-22s matches: %v\n", file, matches)
	}
	// Output:
	// app.log                matches: true
	// important.log          matches: false
	// build/app.js           matches: true
	// build/README.md        matches: false
	// build/docs/api.md      matches: false
	// build/dist/bundle.js   matches: true
}
