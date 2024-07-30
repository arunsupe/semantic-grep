/* contains function to calculate cosine similarity
Also contains a cache to store the similarity values
*/

package similarity

import (
	"math"
)

type SimilarityCache interface {
	MemoizedCalculateSimilarity(queryToken, token string, queryVector, tokenVector interface{}) float64
}

// Implement the interface for your existing cache structure
type Cache struct {
	cache map[string]float64
}

// NewSimilarityCache creates a new cache
func NewSimilarityCache() *Cache {
	return &Cache{
		cache: make(map[string]float64),
	}
}

func (c *Cache) MemoizedCalculateSimilarity(queryToken, token string, queryVector, tokenVector interface{}) float64 {
	// similarity is a commutative operation, we can cache the result for any order of operands
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
