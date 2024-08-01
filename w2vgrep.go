package main

import (
	"fmt"
	"os"

	"w2vgrep/modules/config"
	"w2vgrep/modules/model"
	"w2vgrep/modules/processor"
	"w2vgrep/modules/similarity"

	"github.com/jessevdk/go-flags"
)

// Options defines the command-line options
type Options struct {
	ModelPath           string  `short:"m" long:"model_path" description:"Path to the Word2Vec model file"`
	SimilarityThreshold float64 `short:"t" long:"threshold" default:"0.7" description:"Similarity threshold for matching"`
	ContextBefore       int     `short:"A" long:"before-context" description:"Number of lines before matching line"`
	ContextAfter        int     `short:"B" long:"after-context" description:"Number of lines after matching line"`
	ContextBoth         int     `short:"C" long:"context" description:"Number of lines before and after matching line"`
	PrintLineNumbers    bool    `short:"n" long:"line-number" description:"Print line numbers"`
	IgnoreCase          bool    `short:"i" long:"ignore-case" description:"Ignore case. Note: word2vec is case-sensitive. Ignoring case may lead to unexpected results"`
	OutputOnlyMatching  bool    `short:"o" long:"only-matching" description:"Output only matching words"`
	OutputOnlyLines     bool    `short:"l" long:"only-lines" description:"Output only matched lines without similarity scores"`
}

func main() {
	var opts Options
	var parser = flags.NewParser(&opts, flags.Default)
	parser.Usage = "[OPTIONS] QUERY [FILE]"

	args, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			parser.WriteHelp(os.Stderr)
			os.Exit(1)
		}
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "Error: query is required")
		parser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	if opts.ContextBoth > 0 {
		opts.ContextBefore = opts.ContextBoth
		opts.ContextAfter = opts.ContextBoth
	}

	query := args[0]
	var input *os.File

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

		if opts.ModelPath == "" {
			opts.ModelPath = conf.ModelPath
		}
	}

	if opts.ModelPath == "" {
		fmt.Fprintln(os.Stderr, "Error: Model path is required. Please provide it via config file or -m/--model_path flag.")
		parser.WriteHelp(os.Stderr)
		os.Exit(1)
	}

	var w2vModel model.VectorModel
	var similarityCache similarity.SimilarityCache

	w2vModel, err = model.LoadVectorModel(opts.ModelPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading full model: %v\n", err)
		os.Exit(1)
	}
	similarityCache = similarity.NewSimilarityCache()

	processor.ProcessLineByLine(query, w2vModel, similarityCache, opts.SimilarityThreshold,
		opts.ContextBefore, opts.ContextAfter, input, opts.PrintLineNumbers, opts.IgnoreCase,
		opts.OutputOnlyMatching, opts.OutputOnlyLines)
}
