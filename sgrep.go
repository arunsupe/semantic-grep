package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"strings"
)

const DefaultConfigPath = "config.json"

type Config struct {
	ModelPath string `json:"model_path"`
}

type Word2VecModel struct {
	Vectors map[string][]float32
	Size    int
}

func loadConfig(configPath string) (*Config, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func loadWord2VecModel(modelPath string) (*Word2VecModel, error) {
	file, err := os.Open(modelPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

	// Read header
	var vocabSize, vectorSize int
	fmt.Fscanf(reader, "%d %d\n", &vocabSize, &vectorSize)

	model := &Word2VecModel{
		Vectors: make(map[string][]float32, vocabSize),
		Size:    vectorSize,
	}

	for i := 0; i < vocabSize; i++ {
		word, err := reader.ReadString(' ')
		if err != nil {
			return nil, err
		}
		word = strings.TrimSpace(word)

		vector := make([]float32, vectorSize)
		for j := 0; j < vectorSize; j++ {
			err := binary.Read(reader, binary.LittleEndian, &vector[j])
			if err != nil {
				return nil, err
			}
		}
		model.Vectors[word] = vector
	}

	return model, nil
}

func getVectorEmbedding(token string, model *Word2VecModel) []float32 {
	vec, ok := model.Vectors[token]
	if !ok {
		return make([]float32, model.Size) // Handle OOV words
	}
	return vec
}

func calculateSimilarity(vec1, vec2 []float32) float64 {
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

func colorText(text string, color string) string {
	switch color {
	case "red":
		return fmt.Sprintf("\033[91m%s\033[0m", text)
	case "green":
		return fmt.Sprintf("\033[92m%s\033[0m", text)
	default:
		return text
	}
}

func processLineByLine(query string, model *Word2VecModel, similarityThreshold float64, contextBefore, contextAfter int, input *os.File, printLineNumbers bool, ignoreCase bool) {
	var queryVector []float32
	if ignoreCase {
		queryVector = getVectorEmbedding(strings.ToLower(query), model)
	} else {
		queryVector = getVectorEmbedding(query, model)
	}

	scanner := bufio.NewScanner(input)
	lineNumber := 0
	var contextBuffer []string
	var contextLineNumbers []int

	for scanner.Scan() {
		line := scanner.Text()
		lineNumber++
		tokens := strings.Fields(line)
		matched := false
		var highlightedLine string
		var similarityScore float64

		for _, token := range tokens {
			var tokenToCheck string
			if ignoreCase {
				tokenToCheck = strings.ToLower(token)
			} else {
				tokenToCheck = token
			}

			tokenVector := getVectorEmbedding(tokenToCheck, model)
			similarity := calculateSimilarity(queryVector, tokenVector)
			if similarity > similarityThreshold {
				highlightedLine = strings.Replace(line, token, colorText(token, "red"), -1)
				matched = true
				similarityScore = similarity
				break
			}
		}

		if matched {
			// Print similarity score
			fmt.Printf("Similarity: %.4f\n", similarityScore)

			// Print context before matching line
			for i, ctxLine := range contextBuffer {
				printLine(ctxLine, contextLineNumbers[i], printLineNumbers)
			}
			contextBuffer = nil // Clear context buffer after printing
			contextLineNumbers = nil

			// Print matching line
			printLine(highlightedLine, lineNumber, printLineNumbers)

			// Collect context after matching line
			for i := 0; i < contextAfter && scanner.Scan(); i++ {
				lineNumber++
				printLine(scanner.Text(), lineNumber, printLineNumbers)
			}

			// Print line separator after each match
			fmt.Println("--")
		} else {
			// Maintain context buffer
			if contextBefore > 0 {
				contextBuffer = append(contextBuffer, line)
				contextLineNumbers = append(contextLineNumbers, lineNumber)
				if len(contextBuffer) > contextBefore {
					contextBuffer = contextBuffer[1:]
					contextLineNumbers = contextLineNumbers[1:]
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
	}
}

func printLine(line string, lineNumber int, printLineNumbers bool) {
	if printLineNumbers {
		fmt.Printf("%s:", colorText(fmt.Sprintf("%d", lineNumber), "green"))
	}
	fmt.Println(line)
}

func main() {
	modelPath := flag.String("model_path", "", "Path to the Word2Vec model file")
	similarityThreshold := flag.Float64("threshold", 0.7, "Similarity threshold for matching")
	contextBefore := flag.Int("A", 0, "Number of lines before matching line")
	contextAfter := flag.Int("B", 0, "Number of lines after matching line")
	contextBoth := flag.Int("C", 0, "Number of lines before and after matching line")
	printLineNumbers := flag.Bool("n", false, "Print line numbers")
	ignoreCase := flag.Bool("i", false, "Ignore case. Note: word2vec is case-sensitive. Ignoring case may lead to unexpected results")

	flag.Parse()

	// Override contextBefore and contextAfter if contextBoth is specified
	if *contextBoth > 0 {
		*contextBefore = *contextBoth
		*contextAfter = *contextBoth
	}

	// Remaining arguments are the query and optional filename
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

	// Load configuration from file
	config, err := loadConfig(DefaultConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override modelPath if provided via command line
	if *modelPath != "" {
		config.ModelPath = *modelPath
	}

	model, err := loadWord2VecModel(config.ModelPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading model: %v\n", err)
		os.Exit(1)
	}

	processLineByLine(query, model, *similarityThreshold, *contextBefore, *contextAfter, input, *printLineNumbers, *ignoreCase)

}
