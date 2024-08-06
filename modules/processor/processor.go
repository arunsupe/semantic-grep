// Package processor provides functions to process text line by line,
// performing semantic searches using Word2Vec models.
package processor

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"w2vgrep/modules/model"
	"w2vgrep/modules/similarity"
	"w2vgrep/modules/utils"

	"github.com/clipperhouse/uax29/words"
)

// ProcessLineByLine processes an input file line by line, performing semantic searches
// based on the provided queries and Word2Vec model. It supports various options for
// context lines, case sensitivity, and output formatting.
//
// queries: List of query words to search for.
// w2vModel: The Word2Vec model used for semantic matching.
// similarityCache: Cache for storing similarity calculations.
// similarityThreshold: Threshold above which a match is considered similar.
// contextBefore: Number of lines to include before a matching line.
// contextAfter: Number of lines to include after a matching line.
// input: The input file to process.
// printLineNumbers: Whether to print line numbers in the output.
// ignoreCase: Whether to ignore case when matching words.
// outputOnlyMatching: Whether to output only the matching words.
// outputOnlyLines: Whether to output only the lines that contain matches.
func ProcessLineByLine(queries []string, w2vModel model.VectorModel, similarityCache similarity.SimilarityCache,
	similarityThreshold float64, contextBefore, contextAfter int, input *os.File,
	printLineNumbers, ignoreCase, outputOnlyMatching, outputOnlyLines bool) {

	// Prepare query vectors
	queryVectors := make(map[string]interface{})
	queryInModel := make(map[string]bool)

	for _, query := range queries {
		var queryTokenToCheck string
		if ignoreCase {
			queryTokenToCheck = strings.ToLower(query)
		} else {
			queryTokenToCheck = query
		}

		queryVector, err := w2vModel.GetEmbedding(queryTokenToCheck)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
			queryInModel[queryTokenToCheck] = false
		} else {
			switch queryVector.(type) {
			case []float32, []int8:
				queryVectors[queryTokenToCheck] = queryVector
				queryInModel[queryTokenToCheck] = true
			default:
				fmt.Fprintf(os.Stderr, "Warning: Unsupported vector type for query: %s\n", queryTokenToCheck)
				queryInModel[queryTokenToCheck] = false
			}
		}
	}

	scanner := bufio.NewScanner(input)
	lineNumber := 0
	var contextBuffer []string
	var contextLineNumbers []int

	// Process each line
	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++
		matched := false
		var highlightedLine string
		var similarityScore float64
		var matchSimilarityScore float64

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

			for queryTokenToCheck, queryVector := range queryVectors {
				// Check if tokenToCheck is exactly equal to queryTokenToCheck
				if tokenToCheck == queryTokenToCheck {
					similarityScore = 1.0
					matchSimilarityScore = similarityScore
					matched = true
					highlightedLine = strings.Replace(line, token, utils.ColorText(token, "red"), -1)
				} else if queryInModel[queryTokenToCheck] {
					// Only perform similarity check if query is in the model
					tokenVector, err := w2vModel.GetEmbedding(tokenToCheck)
					if err == nil {
						// Calculate similarity and check threshold only if token is in model
						similarityScore = similarityCache.MemoizedCalculateSimilarity(queryTokenToCheck, tokenToCheck, queryVector, tokenVector)
						if similarityScore > similarityThreshold {
							matched = true
							highlightedLine = strings.Replace(line, token, utils.ColorText(token, "red"), -1)
							matchSimilarityScore = similarityScore
						}
					}
				}

				if matched {
					if outputOnlyMatching {
						fmt.Println(token)
						break // Stop after first match if -o is set
					}
					break // Stop checking other tokens in this line
				}
			}
		}

		// Handle matched line
		if matched {
			if outputOnlyMatching {
				// Already printed in the loop above
			} else if outputOnlyLines {
				utils.PrintLine(highlightedLine, lineNumber, printLineNumbers)
			} else {
				fmt.Printf("Similarity: %.4f\n", matchSimilarityScore)
				// Print the context lines before the match
				for i, ctxLine := range contextBuffer {
					utils.PrintLine(ctxLine, contextLineNumbers[i], printLineNumbers)
				}

				// Print the matched line with highlighted token
				utils.PrintLine(highlightedLine, lineNumber, printLineNumbers)

				// Print the context lines after the match
				for i := 0; i < contextAfter && scanner.Scan(); i++ {
					lineNumber++
					utils.PrintLine(scanner.Text(), lineNumber, printLineNumbers)
				}

				fmt.Println("--")
			}

			// Clear the context buffer after printing
			contextBuffer = nil
			contextLineNumbers = nil
		} else {
			// Update the context buffer with the current line if no match is found
			if contextBefore > 0 && !outputOnlyMatching && !outputOnlyLines {
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
