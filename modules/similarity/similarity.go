/* contains function to calculate cosine similarity
Also contains a cache to store the similarity values
*/

package similarity

import (
	"math"
)

// SimilarityCache is a cache for storing similarity values between tokens
type SimilarityCache struct {
	cache map[string]float64
}

// NewSimilarityCache creates a new SimilarityCache
func NewSimilarityCache() *SimilarityCache {
	return &SimilarityCache{
		cache: make(map[string]float64),
	}
}

// MemoizedCalculateSimilarity calculates the similarity between two tokens
func (sc *SimilarityCache) MemoizedCalculateSimilarity(queryToken, token string, queryVector, tokenVector []float32) float64 {
	// Create a key for the cache by concatenating the query and input tokens
	key := queryToken + "|" + token

	// If the key is too long, don't cache the result
	if len(key) > 30 {
		return calculateSimilarity(queryVector, tokenVector)
	}

	// Check if the result is already in the cache
	if result, found := sc.cache[key]; found {
		return result
	}

	// Calculate the similarity and store it in the cache
	result := calculateSimilarity(queryVector, tokenVector)
	sc.cache[key] = result

	return result
}

// calculateSimilarity calculates the cosine similarity between two vectors
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
