// Package similarity provides functions and types for calculating and caching
// the similarity between word vectors using cosine similarity.
package similarity

import (
	"math"
)

// SimilarityCache is an interface for caching and calculating the similarity
// between word vectors.
type SimilarityCache interface {
	// MemoizedCalculateSimilarity calculates the similarity between two word vectors
	// and caches the result to avoid redundant calculations.
	MemoizedCalculateSimilarity(queryToken, token string, queryVector, tokenVector interface{}) float64
}

// Cache implements the SimilarityCache interface and provides a simple in-memory cache.
type Cache struct {
	cache map[string]float64
}

// NewSimilarityCache creates a new Cache instance for storing similarity calculations.
func NewSimilarityCache() *Cache {
	return &Cache{
		cache: make(map[string]float64),
	}
}

// MemoizedCalculateSimilarity calculates the similarity between two word vectors
// and caches the result. It supports both []float32 and []int8 vector types.
func (c *Cache) MemoizedCalculateSimilarity(queryToken, token string, queryVector, tokenVector interface{}) float64 {
	key := token

	if cachedValue, exists := c.cache[key]; exists {
		return cachedValue
	}

	var similarity float64
	switch qv := queryVector.(type) {
	case []float32:
		similarity = calculateSimilarity32bit(qv, tokenVector.([]float32))
	case []int8:
		similarity = calculateSimilarity8bit(qv, tokenVector.([]int8))
	default:
		panic("Unsupported vector type")
	}

	c.cache[key] = similarity
	return similarity
}

// calculateSimilarity calculates the cosine similarity between two []float32 vectors
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

// calculateSimilarity calculates the cosine similarity between two []int8 vectors
func calculateSimilarity8bit(vec1, vec2 []int8) float64 {
	var dotProduct int32
	var norm1, norm2 int32

	for i := range vec1 {
		dotProduct += int32(vec1[i]) * int32(vec2[i])
		norm1 += int32(vec1[i]) * int32(vec1[i])
		norm2 += int32(vec2[i]) * int32(vec2[i])
	}

	return float64(dotProduct) / (math.Sqrt(float64(norm1)) * math.Sqrt(float64(norm2)))
}
