/*
Package processor provides functionality to process input files line by line.

It embeds the query, tokenizes input lines, embeds each token, and calculates
the similarity between the query and each token in the line. If the similarity
is above a threshold, it prints the line with the matched token highlighted.

Key features:
- Case sensitivity handling
- Context printing (before and after matching lines)
- Line number printing
- Similarity caching for performance
*/

package processor

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"sgrep/modules/model"
	"sgrep/modules/similarity"
	"sgrep/modules/utils"

	"github.com/clipperhouse/uax29/words"
)

// ProcessLineByLine processes the input file line by line, comparing each token
// to the query and printing matches with context
func ProcessLineByLine(query string, w2vModel *model.Word2VecModel, similarityThreshold float64, contextBefore, contextAfter int, input *os.File, printLineNumbers bool, ignoreCase bool) {
	// Prepare query vector
	var queryVector []float32
	var queryTokenToCheck string
	if ignoreCase {
		queryTokenToCheck = strings.ToLower(query)
	} else {
		queryTokenToCheck = query
	}
	queryVector = model.GetVectorEmbedding(queryTokenToCheck, w2vModel)

	scanner := bufio.NewScanner(input)
	lineNumber := 0
	var contextBuffer []string
	var contextLineNumbers []int

	similarityCache := similarity.NewSimilarityCache()

	// Process each line
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++
		matched := false
		var highlightedLine string
		var similarityScore float64

		// Tokenize and check each token
		tokens := words.NewSegmenter(scanner.Bytes())
		for tokens.Next() {
			token := tokens.Text()
			var tokenToCheck string
			if ignoreCase {
				tokenToCheck = strings.ToLower(token)
			} else {
				tokenToCheck = token
			}

			// Calculate similarity and check threshold
			tokenVector := model.GetVectorEmbedding(tokenToCheck, w2vModel)
			similarity := similarityCache.MemoizedCalculateSimilarity(queryTokenToCheck, tokenToCheck, queryVector, tokenVector)
			if similarity > similarityThreshold {
				highlightedLine = strings.Replace(line, token, utils.ColorText(token, "red"), -1)
				matched = true
				similarityScore = similarity
				break
			}
		}

		// Handle matched line
		if matched {
			fmt.Printf("Similarity: %.4f\n", similarityScore)
			// Print the context lines before the match
			for i, ctxLine := range contextBuffer {
				utils.PrintLine(ctxLine, contextLineNumbers[i], printLineNumbers)
			}
			// Clear the context buffer as it has been printed
			contextBuffer = nil
			contextLineNumbers = nil

			// Print the matched line with highlighted token
			utils.PrintLine(highlightedLine, lineNumber, printLineNumbers)

			// Print the context lines after the match
			for i := 0; i < contextAfter && scanner.Scan(); i++ {
				lineNumber++
				utils.PrintLine(scanner.Text(), lineNumber, printLineNumbers)
			}

			fmt.Println("--")
		} else {
			// Update the context buffer with the current line if no match is found
			if contextBefore > 0 {
				contextBuffer = append(contextBuffer, line)
				contextLineNumbers = append(contextLineNumbers, lineNumber)
				// Ensure the context buffer does not exceed the specified number of lines
				if len(contextBuffer) > contextBefore {
					contextBuffer = contextBuffer[1:]
					contextLineNumbers = contextLineNumbers[1:]
				}
			}
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}
