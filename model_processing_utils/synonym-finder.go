package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// Options defines the command-line options
type Options struct {
	ModelPath           string
	SimilarityThreshold float64
	IgnoreCase          bool
	PatternFile         string
	OnlyMatching        bool // New field for -o flag
}

// VectorModel interface defines the methods that all vector models must implement
type VectorModel interface {
	LoadModel(filename string) error
	GetEmbedding(token string) interface{}
}

// VecModel32bit represents a 32-bit floating point Word2Vec model
type VecModel32bit struct {
	Vectors map[string][]float32
	Size    int
}

// LoadModel loads a 32-bit floating point Word2Vec model from a file
func (m *VecModel32bit) LoadModel(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Read header
	var vocabSize, vectorSize int
	_, err = fmt.Fscanf(reader, "%d %d\n", &vocabSize, &vectorSize)
	if err != nil {
		return fmt.Errorf("failed to read header: %v\nCheck that you have a valid model file", err)
	}

	// Validate header
	if vocabSize <= 0 || vectorSize <= 0 {
		return fmt.Errorf("invalid header: vocabSize=%d, vectorSize=%d\nCheck that you have a valid model file", vocabSize, vectorSize)
	}

	m.Vectors = make(map[string][]float32, vocabSize)
	m.Size = vectorSize

	for i := 0; i < vocabSize; i++ {
		word, err := reader.ReadString(' ')
		if err != nil {
			return fmt.Errorf("failed to read word: %v", err)
		}
		word = strings.TrimSpace(word)

		vector := make([]float32, vectorSize)
		for j := 0; j < vectorSize; j++ {
			err := binary.Read(reader, binary.LittleEndian, &vector[j])
			if err != nil {
				return fmt.Errorf("failed to read vector: %v", err)
			}
		}

		// Check if we've reached the end of the record
		nextByte, err := reader.Peek(1)
		if err != nil && err != io.EOF {
			return fmt.Errorf("unexpected error reading next byte: %v", err)
		}
		if len(nextByte) > 0 && nextByte[0] == '\n' {
			reader.ReadByte() // consume the newline
		}

		m.Vectors[word] = vector
	}

	// Check if we've reached the end of the file
	_, err = reader.ReadByte()
	if err != io.EOF {
		return fmt.Errorf("unexpected data at end of file.\nCheck that you have a valid model file")
	}

	return nil
}

// GetEmbedding returns the vector embedding of a token for the 32-bit model
func (m *VecModel32bit) GetEmbedding(token string) interface{} {
	vec, ok := m.Vectors[token]
	if !ok {
		return make([]float32, m.Size)
	}
	return vec
}

// LoadVectorModel loads either a 32-bit or 8-bit model based on the file extension
func LoadVectorModel(filename string) (VectorModel, error) {
	var model VectorModel

	if strings.HasSuffix(filename, ".bin") {
		model = &VecModel32bit{}
	} else {
		return nil, fmt.Errorf("unsupported file format")
	}

	err := model.LoadModel(filename)
	if err != nil {
		return nil, err
	}

	return model, nil
}

// calculateSimilarity calculates the cosine similarity between two vectors
func calculateSimilarity32bit(vec1, vec2 []float32) float64 {
	dotProduct := float64(0)
	norm1 := float64(0)
	norm2 := float64(0)
	for i := range vec1 {
		dotProduct += float64(vec1[i] * vec2[i])
		norm1 += float64(vec1[i] * vec1[i])
		norm2 += float64(vec2[i] * vec2[i])
	}
	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// findSimilarWords finds words in the model that are similar to the query word above the given threshold
func findSimilarWords(model VectorModel, query string, threshold float64, onlyMatching bool) error {
	queryEmbedding := model.GetEmbedding(query).([]float32)
	if len(queryEmbedding) == 0 {
		return fmt.Errorf("query word not found in model")
	}

	if onlyMatching {
		fmt.Println(query) // Print the bare query
	} else {
		fmt.Printf("Words similar to '%s' with similarity >= %.2f:\n", query, threshold)
	}

	for word, embedding := range model.(*VecModel32bit).Vectors {
		similarity := calculateSimilarity32bit(queryEmbedding, embedding)
		if similarity >= threshold && similarity < 1.0 {
			if onlyMatching {
				fmt.Println(word)
			} else {
				fmt.Printf("%s %.4f\n", word, similarity)
			}
		}
	}

	return nil
}

// findSimilarWordsForPatterns finds similar words for each pattern in the given file
func findSimilarWordsForPatterns(model VectorModel, patternFile string, threshold float64, onlyMatching bool) error {
	file, err := os.Open(patternFile)
	if err != nil {
		return fmt.Errorf("failed to open pattern file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		pattern := strings.TrimSpace(scanner.Text())
		if pattern == "" {
			continue
		}

		err := findSimilarWords(model, pattern, threshold, onlyMatching)
		if err != nil {
			fmt.Printf("Warning: %v\n", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading pattern file: %v", err)
	}

	return nil
}

func main() {
	var opts Options

	flag.StringVar(&opts.ModelPath, "model_path", "", "Path to the Word2Vec model file (required)")
	flag.Float64Var(&opts.SimilarityThreshold, "threshold", 0.7, "Similarity threshold for matching (default 0.7)")
	flag.BoolVar(&opts.IgnoreCase, "ignore-case", false, "Ignore case. Note: word2vec is case-sensitive. Ignoring case may lead to unexpected results")
	flag.StringVar(&opts.PatternFile, "f", "", "File containing patterns, one per line")
	flag.BoolVar(&opts.OnlyMatching, "o", false, "Print only matching tokens")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] [QUERY]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "QUERY is the word to find similar words for (required if -f is not used)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -model_path path/to/model.bin -threshold 0.8 cat\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -model_path path/to/model.bin -threshold 0.8 -f patterns.txt\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -model_path path/to/model.bin -threshold 0.8 -o cat\n", os.Args[0])
	}

	flag.Parse()

	if opts.ModelPath == "" {
		fmt.Fprintln(os.Stderr, "Error: Model path is required. Please provide it via -model_path flag.")
		flag.Usage()
		os.Exit(1)
	}

	if flag.Lookup("threshold").Value.String() == "0.7" {
		fmt.Fprintln(os.Stderr, "Error: Threshold is required. Please provide it via -threshold flag.")
		flag.Usage()
		os.Exit(1)
	}

	model, err := LoadVectorModel(opts.ModelPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading model: %v\n", err)
		os.Exit(1)
	}

	if opts.PatternFile != "" {
		err = findSimilarWordsForPatterns(model, opts.PatternFile, opts.SimilarityThreshold, opts.OnlyMatching)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error processing pattern file: %v\n", err)
			os.Exit(1)
		}
	} else {
		args := flag.Args()
		if len(args) != 1 {
			fmt.Fprintln(os.Stderr, "Error: Exactly one query word is required when not using -f")
			flag.Usage()
			os.Exit(1)
		}

		query := args[0]
		err = findSimilarWords(model, query, opts.SimilarityThreshold, opts.OnlyMatching)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding similar words: %v\n", err)
			os.Exit(1)
		}
	}
}
