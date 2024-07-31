/* Process command line arguments and call the processor to
process the input file line by line.
*/

package main

import (
	"flag"
	"fmt"
	"os"

	"sgrep/modules/config"
	"sgrep/modules/model"
	"sgrep/modules/processor"
	"sgrep/modules/similarity"
)

func main() {
	modelPath := flag.String("model_path", "", "Path to the Word2Vec model file")
	similarityThreshold := flag.Float64("threshold", 0.7, "Similarity threshold for matching")
	contextBefore := flag.Int("A", 0, "Number of lines before matching line")
	contextAfter := flag.Int("B", 0, "Number of lines after matching line")
	contextBoth := flag.Int("C", 0, "Number of lines before and after matching line")
	printLineNumbers := flag.Bool("n", false, "Print line numbers")
	ignoreCase := flag.Bool("i", false, "Ignore case. Note: word2vec is case-sensitive. Ignoring case may lead to unexpected results")

	flag.Parse()

	if *contextBoth > 0 {
		*contextBefore = *contextBoth
		*contextAfter = *contextBoth
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: query is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	query := args[0]
	var input *os.File
	var err error

	if len(args) > 1 {
		input, err = os.Open(args[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
			os.Exit(1)
		}
		defer input.Close()
	} else {
		input = os.Stdin
	}

	configPath := config.FindConfigFile()

	if configPath != "" {
		conf, err := config.LoadConfig(configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config from %s: %v\n", configPath, err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Using configuration file: %s\n", configPath)

		if *modelPath == "" {
			*modelPath = conf.ModelPath
		}
	}

	if *modelPath == "" {
		fmt.Fprintln(os.Stderr, "Error: Model path is required. Please provide it via config file or -model_path flag.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var w2vModel model.VectorModel
	var similarityCache similarity.SimilarityCache

	w2vModel, err = model.LoadVectorModel(*modelPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading full model: %v\n", err)
		os.Exit(1)
	}
	similarityCache = similarity.NewSimilarityCache()

	// Dereference the pointers when passing to ProcessLineByLine
	processor.ProcessLineByLine(query, w2vModel, similarityCache, *similarityThreshold, *contextBefore, *contextAfter, input, *printLineNumbers, *ignoreCase)

}
