package utils

import (
	"fmt"
)

// ColorText colors the given text with the specified color.
func ColorText(text, color string) string {
	colors := map[string]string{
		"red":     "\033[31m",
		"green":   "\033[32m",
		"yellow":  "\033[33m",
		"blue":    "\033[34m",
		"magenta": "\033[35m",
		"cyan":    "\033[36m",
		"reset":   "\033[0m",
	}
	return colors[color] + text + colors["reset"]
}

// PrintLine prints a line with an optional line number.
func PrintLine(line string, lineNumber int, printLineNumbers bool) {
	if printLineNumbers {
		lineNumberStr := ColorText(fmt.Sprintf("%d:", lineNumber), "blue")
		fmt.Printf("%s %s\n", lineNumberStr, line)
	} else {
		fmt.Println(line)
	}
}
