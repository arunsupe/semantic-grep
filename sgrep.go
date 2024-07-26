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

// Configuration
const (
	DefaultConfigPath = "config.json"
)

var stopWords = map[string]bool{
	"of": true, "to": true, "in": true, "on": true, "at": true, "by": true, "with": true,
	"the": true, "a": true, "an": true, "and": true, "but": true, "or": true, "so": true,
	"yet": true, "is": true, "are": true, "am": true, "be": true, "been": true, "not": true,
	"only": true, "just": true, "still": true, "already": true, "I": true, "me": true,
	"my": true, "mine": true, "you": true, "your": true, "yours": true, "he": true,
	"him": true, "his": true, "it": true, "its": true, "they": true, "them": true,
	"their": true, "theirs": true, "oh": true, "ah": true, "yes": true, "no": true,
	"okay": true, "one": true, "two": true, "three": true, "four": true, "five": true,
	"six": true, "seven": true, "eight": true, "nine": true,
}

type Config struct {
	ModelPath           string  `json:"model_path"`
	SimilarityThreshold float64 `json:"similarity_threshold"`
	Window              int     `json:"window"`
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

func toVector(tokens []string, model *Word2VecModel) []float32 {
	sum := make([]float32, model.Size)
	count := 0
	for _, token := range tokens {
		token = strings.ToLower(token)
		if !stopWords[token] {
			vec := getVectorEmbedding(token, model)
			for i, v := range vec {
				sum[i] += v
			}
			count++
		}
	}
	if count == 0 {
		return sum
	}
	for i := range sum {
		sum[i] /= float32(count)
	}
	return sum
}

func processTokenStream(query string, model *Word2VecModel, window int, similarityThreshold float64) {
	queryVector := toVector(strings.Fields(query), model)
	tokenBuffer := make([]string, 0, window)
	stepSize := window / 10

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		token := scanner.Text()
		tokenBuffer = append(tokenBuffer, token)

		if len(tokenBuffer) == window {
			if countPunctuations(tokenBuffer) <= int(0.2*float64(window)) {
				tokenVector := toVector(tokenBuffer, model)
				similarity := calculateSimilarity(queryVector, tokenVector)

				if similarity > similarityThreshold {
					fmt.Printf("Similarity: %.4f\n", similarity)
					colorText(strings.Join(tokenBuffer, " "), int(similarity*5)+1)
					fmt.Println()
				}
			}

			// Move the window by step_size
			tokenBuffer = tokenBuffer[stepSize:]
		}
	}
}

func calculateSimilarity(queryVector, tokenVector []float32) float64 {
	dotProduct := float64(0)
	normQuery := float64(0)
	normToken := float64(0)
	for i := range queryVector {
		dotProduct += float64(queryVector[i] * tokenVector[i])
		normQuery += float64(queryVector[i] * queryVector[i])
		normToken += float64(tokenVector[i] * tokenVector[i])
	}
	return dotProduct / (math.Sqrt(normQuery) * math.Sqrt(normToken))
}

func countPunctuations(textList []string) int {
	count := 0
	for _, text := range textList {
		for _, char := range text {
			if strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", char) {
				count++
			}
		}
	}
	return count
}

func colorText(text string, importanceLevel int) {
	colors := map[int]string{
		1: "\033[91m", // Bright Red
		2: "\033[93m", // Bright Yellow
		3: "\033[94m", // Bright Blue
		4: "\033[95m", // Bright Magenta
		5: "\033[0m",  // Default color
	}
	fmt.Printf("%s%s\033[0m\n", colors[importanceLevel], text)
}

func main() {
	fmt.Println("sgrep - Semantic Grep")

	query := flag.String("query", "", "Query string for semantic search")
	configPath := flag.String("config", DefaultConfigPath, "Path to the configuration file")
	modelPath := flag.String("model_path", "", "Path to the Word2Vec model file")
	similarityThreshold := flag.Float64("similarity_threshold", 0, "Similarity threshold for matching")
	window := flag.Int("window", 0, "Window size for token stream")

	flag.Parse()

	// Load configuration from file
	config, err := loadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Override config values with CLI values if provided
	if *modelPath != "" {
		config.ModelPath = *modelPath
	}
	if *similarityThreshold != 0 {
		config.SimilarityThreshold = *similarityThreshold
	}
	if *window != 0 {
		config.Window = *window
	}

	if *query == "" {
		fmt.Fprintln(os.Stderr, "Error: query is required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	model, err := loadWord2VecModel(config.ModelPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading model: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Query: %s\n", *query)

	processTokenStream(*query, model, config.Window, config.SimilarityThreshold)
}
