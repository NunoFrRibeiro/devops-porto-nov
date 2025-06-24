package main

import (
	"fmt"
	"regexp"
	"strconv"
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

	lines := strings.Split(diffText, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "---") {
			continue
		} else if strings.HasPrefix(line, "+++") {
			currentFile = strings.TrimPrefix(line, "+++ /b/")
		} else if strings.HasPrefix(line, "@@") {
			re := regexp.MustCompile(`@@ -\d+,\d+ \+(\d+),\d+ @@`)
			match := re.FindStringSubmatch(line)
			if len(match) > 1 {
				lineNumber, err := strconv.Atoi(match[1])
				if err != nil {
					fmt.Println("Error converting line number:", err)
					continue
				}
				currentLine = lineNumber
				newCode = []string{}
				removalReached = false
			}

		} else if strings.HasPrefix(line, "-") {
			removalReached = true
		} else if strings.HasPrefix(line, "+") {
			if currentFile != "" && currentLine != 0 {
				newCode = append(newCode, strings.TrimPrefix(line, "+"))
			}
		} else {
			if removalReached && len(newCode) > 0 && currentFile != "" && currentLine != 0 {
				suggestions = append(suggestions, CodeSuggestion{
					File:       currentFile,
					Line:       currentLine,
					Suggestion: newCode,
				})
				newCode = []string{}
				removalReached = false
			}
			currentLine++
		}
	}

	if removalReached && len(newCode) > 0 && currentFile != "" && currentLine != 0 {
		suggestions = append(suggestions, CodeSuggestion{
			File:       currentFile,
			Line:       currentLine,
			Suggestion: newCode,
		})
	}

	return suggestions
}
