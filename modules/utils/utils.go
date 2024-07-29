/* Functions for colorizing text and printing lines with line numbers
 */
package utils

import (
	"fmt"
)

// ColorText returns the text with the specified color
func ColorText(text string, color string) string {
	switch color {
	case "red":
		return fmt.Sprintf("\033[91m%s\033[0m", text)
	case "green":
		return fmt.Sprintf("\033[92m%s\033[0m", text)
	default:
		return text
	}
}

// PrintLine prints a line with an optional line number
func PrintLine(line string, lineNumber int, printLineNumbers bool) {
	if printLineNumbers {
		fmt.Printf("%s:", ColorText(fmt.Sprintf("%d", lineNumber), "green"))
	}
	fmt.Println(line)
}
