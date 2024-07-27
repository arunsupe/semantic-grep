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

/*
processTokenStream processes the input text stream and finds semantically similar chunks to the given query.
The function follows these steps:

1. Convert the query to a vector representation.
2. Initialize a token buffer and a scanner to read input tokens.
3. For each token in the input stream:
   a. Add the token to the buffer.
   b. When the buffer reaches the window size:
      - Check if the chunk has an acceptable amount of punctuation.
      - Convert the chunk to a vector and calculate its similarity to the query.
      - If the similarity is above the threshold and higher than the current best:
        * Update the best chunk and its similarity score.
      - Increment the count of processed tokens.
      - If a full window's worth of tokens has been processed:
        * Print the best chunk found (if any).
        * Reset the best similarity and processed token count.
      - Slide the window forward by removing the oldest tokens.
4. After processing all tokens, print any remaining best chunk.

This approach ensures that:
- Only non-overlapping chunks are printed.
- Each printed chunk is the most similar to the query within its window.
- The sliding window mechanism is maintained, allowing for flexible matching.
- Output is significantly reduced compared to printing every similar chunk.
*/
func processTokenStream(query string, model *Word2VecModel, window int, similarityThreshold float64) {
	// Convert the query to a vector representation
	queryVector := toVector(strings.Fields(query), model)

	// Initialize a buffer to hold tokens, with capacity equal to the window size
	tokenBuffer := make([]string, 0, window)

	// Calculate the step size for sliding the window
	// This determines how many tokens we move forward after processing each window
	stepSize := window / 10

	// Set up a scanner to read input token by token
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Split(bufio.ScanWords)

	// Variables to keep track of the best matching chunk within a non-overlapping window
	bestChunk := make([]string, 0, window)
	bestSimilarity := 0.0
	tokensProcessed := 0

	// Main loop to process input tokens
	for scanner.Scan() {
		// Read the next token
		token := scanner.Text()

		// Add the token to our buffer
		tokenBuffer = append(tokenBuffer, token)

		// Process the buffer when it reaches the window size
		if len(tokenBuffer) == window {
			// Check if the chunk doesn't have too many punctuations
			if countPunctuations(tokenBuffer) <= int(0.2*float64(window)) {
				// Convert the current chunk to a vector
				tokenVector := toVector(tokenBuffer, model)

				// Calculate similarity between the chunk and the query
				similarity := calculateSimilarity(queryVector, tokenVector)

				// If this chunk is more similar than the previous best and above threshold,
				// update our best chunk
				if similarity > similarityThreshold && similarity > bestSimilarity {
					bestSimilarity = similarity
					bestChunk = make([]string, len(tokenBuffer))
					copy(bestChunk, tokenBuffer)
				}
			}

			// Increment our count of processed tokens
			tokensProcessed += stepSize

			// If we've processed a full window's worth of tokens,
			// print the best chunk we've found (if any)
			if tokensProcessed >= window {
				if bestSimilarity > 0 {
					fmt.Printf("Similarity: %.4f\n", bestSimilarity)
					colorText(strings.Join(bestChunk, " "), int(bestSimilarity*5)+1)
					fmt.Println()
				}
				// Reset our tracking variables for the next window
				bestSimilarity = 0
				tokensProcessed = 0
			}

			// Slide the window forward by removing processed tokens
			tokenBuffer = tokenBuffer[stepSize:]
		}
	}

	// After processing all input, check if there's a remaining best chunk to print
	// This handles cases where the input ends before completing another full window
	if bestSimilarity > 0 {
		fmt.Printf("Similarity: %.4f\n", bestSimilarity)
		colorText(strings.Join(bestChunk, " "), int(bestSimilarity*5)+1)
		fmt.Println()
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
