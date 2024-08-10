/*
Description:
This program clusters words into synonyms, printing the clusters to a file,
one cluster per line. It's output can be used by standard grep to find synonyms
in texts. Some of the clusters are too big for grep to handle; grep -f may work.
Not sure if this tool is useful. But an INTERESTING EXPERIMENT.

This script is used to cluster the words in the word2vec model using
mini-batch k-means clustering. It works for any language as long as the model is
in that language. The output is a text file where each line contains a cluster of
synonyms separated by a pipe (|) character. I want to use it in the
standard unix text tools.

The script takes the path to the word2vec binary model, the number of clusters (k),
the batch size for mini-batch k-means, the maximum number of iterations, and
the output file path as input.
The script performs mini-batch k-means clustering on the word vectors and
writes the clusters to the output file.

Usage: cluster.go -model path/to/model.bin \
				-k 100 -batch-size 100 \
				-iterations 100 \
				-output clusters.txt
*/

package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"
)

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

// func euclideanDistance(vec1, vec2 []float32) float64 {
// 	sum := float64(0)
// 	for i := range vec1 {
// 		diff := float64(vec1[i] - vec2[i])
// 		sum += diff * diff
// 	}
// 	return math.Sqrt(sum)
// }

// cosineSimilarity calculates the cosine similarity between two vectors
func cosineSimilarity(vec1, vec2 []float32) float64 {
	dotProduct := float64(0)
	normVec1 := float64(0)
	normVec2 := float64(0)
	for i := range vec1 {
		dotProduct += float64(vec1[i] * vec2[i])
		normVec1 += float64(vec1[i] * vec1[i])
		normVec2 += float64(vec2[i] * vec2[i])
	}
	if normVec1 == 0 || normVec2 == 0 {
		return 0 // To handle zero vectors
	}
	return dotProduct / (math.Sqrt(normVec1) * math.Sqrt(normVec2))
}

// cosineDistance converts cosine similarity to a distance metric
func cosineDistance(vec1, vec2 []float32) float64 {
	return 1 - cosineSimilarity(vec1, vec2)
}

func calculateCentroid(vectors [][]float32) []float32 {
	dim := len(vectors[0])
	centroid := make([]float32, dim)
	for _, vec := range vectors {
		for i := range vec {
			centroid[i] += vec[i]
		}
	}
	for i := range centroid {
		centroid[i] /= float32(len(vectors))
	}
	return centroid
}

// Use cosineDistance
func miniBatchKMeans(vectors [][]float32, words []string, k, batchSize, maxIterations int) [][]string {
	rand.Seed(time.Now().UnixNano())
	dim := len(vectors[0])

	// Initialize k random centroids
	centroids := make([][]float32, k)
	for i := range centroids {
		centroids[i] = vectors[rand.Intn(len(vectors))]
	}

	for iteration := 0; iteration < maxIterations; iteration++ {
		// Sample a random batch of data points
		batchIndices := rand.Perm(len(vectors))[:batchSize]
		batch := make([][]float32, batchSize)
		for i, idx := range batchIndices {
			batch[i] = vectors[idx]
		}

		// Assign points in the batch to the nearest centroid
		clusterAssignments := make([]int, batchSize)
		for i, vec := range batch {
			bestCluster := 0
			// bestDistance := euclideanDistance(vec, centroids[0])
			bestDistance := cosineDistance(vec, centroids[0])
			for j := 1; j < k; j++ {
				// distance := euclideanDistance(vec, centroids[j])
				distance := cosineDistance(vec, centroids[j])
				if distance < bestDistance {
					bestDistance = distance
					bestCluster = j
				}
			}
			clusterAssignments[i] = bestCluster
		}

		// Update centroids based on the batch
		clusterSums := make([][]float32, k)
		clusterCounts := make([]int, k)
		for i := range clusterSums {
			clusterSums[i] = make([]float32, dim)
		}
		for i, vec := range batch {
			cluster := clusterAssignments[i]
			for j := range vec {
				clusterSums[cluster][j] += vec[j]
			}
			clusterCounts[cluster]++
		}
		for i := range centroids {
			if clusterCounts[i] > 0 {
				for j := range centroids[i] {
					centroids[i][j] = clusterSums[i][j] / float32(clusterCounts[i])
				}
			}
		}
	}

	// Assign all points to the nearest centroid
	clusters := make([][]string, k)
	for i, vec := range vectors {
		bestCluster := 0
		bestDistance := cosineDistance(vec, centroids[0])
		for j := 1; j < k; j++ {
			distance := cosineDistance(vec, centroids[j])
			if distance < bestDistance {
				bestDistance = distance
				bestCluster = j
			}
		}
		clusters[bestCluster] = append(clusters[bestCluster], words[i])
	}

	return clusters
}

func main() {
	modelPath := flag.String("model", "", "Path to word2vec binary model")
	k := flag.Int("k", 100, "Number of clusters")
	batchSize := flag.Int("batch-size", 100, "Batch size for mini-batch k-means")
	maxIterations := flag.Int("iterations", 100, "Maximum number of iterations for mini-batch k-means")
	outputPath := flag.String("output", "clusters.txt", "Output file path")
	flag.Parse()

	if *modelPath == "" {
		log.Fatal("Please provide a path to the word2vec binary model")
	}

	// Load the word2vec model
	model, err := LoadVectorModel(*modelPath)
	if err != nil {
		log.Fatalf("Failed to load model: %v", err)
	}

	// Get all words and vectors
	var words []string
	var vectors [][]float32
	for word, vec := range model.(*VecModel32bit).Vectors {
		words = append(words, word)
		vectors = append(vectors, vec)
	}

	// Perform mini-batch k-means clustering
	clusters := miniBatchKMeans(vectors, words, *k, *batchSize, *maxIterations)

	// Sort clusters by size (largest first)
	sort.Slice(clusters, func(i, j int) bool {
		return len(clusters[i]) > len(clusters[j])
	})

	// Write clusters to file
	file, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, cluster := range clusters {
		_, err := writer.WriteString(strings.Join(cluster, "|") + "\n")
		if err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
	}
	writer.Flush()

	fmt.Printf("Clustering complete. %d clusters written to %s\n", len(clusters), *outputPath)
}
