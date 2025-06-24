package main

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
)

type CodeSuggestion struct {
	File       string
	Line       int
	Suggestion []string
}

func parseDiff(diffText string) []CodeSuggestion {
	var suggestions []CodeSuggestion
	var currentFile string
	var currentLine int
	var newCode []string
	removalReached := false

	fileRegex := regexp.MustCompile(`^\+\+\+ /?b/(.+)`)
	lineRegex := regexp.MustCompile(`^@@ .* \+(\d+),?`)

	scanner := bufio.NewScanner(strings.NewReader(diffText))
	for scanner.Scan() {
		line := scanner.Text()

		if matches := fileRegex.FindStringSubmatch(line); matches != nil {
			currentFile = matches[1]
			continue
		}

		if matches := lineRegex.FindStringSubmatch(line); matches != nil {
			currentLine = atoi(matches[1]) - 1
			newCode = []string{}
			removalReached = false
			continue
		}

		if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			newCode = append(newCode, line[1:]) // Remove `+`
			continue
		}

		if !removalReached {
			currentLine++
		}

		if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			if len(newCode) > 0 && currentFile != "" {
				suggestions = append(suggestions, CodeSuggestion{
					File:       currentFile,
					Line:       currentLine,
					Suggestion: newCode,
				})
				newCode = []string{}
			}
			removalReached = true
		}

	}

	if len(newCode) > 0 && currentFile != "" {
		suggestions = append(suggestions, CodeSuggestion{
			File:       currentFile,
			Line:       currentLine,
			Suggestion: newCode,
		})
	}

	return suggestions
}

func atoi(s string) int {
	var i int
	fmt.Sscanf(s, "%d", &i)
	return i
}

func determineProjectBasePath(filenameFromDiff string) string {
	if strings.Contains(strings.ToLower(filenameFromDiff), "counter") {
		return "CounterBackend" // No trailing slash needed here for filepath.Join
	}
	if strings.Contains(strings.ToLower(filenameFromDiff), "adder") {
		return "AdderBackend"
	}

	fmt.Printf(
		"Warning: Could not determine project base path for file '%s'. Assuming repo root.\n",
		filenameFromDiff,
	)
	return ""
}
