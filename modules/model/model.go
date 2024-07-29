/* Package model provides a Word2VecModel struct, a function
to load a Word2Vec model from a file and a function to get
the vector embedding of a token from the model.
*/
package model

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

type Word2VecModel struct {
	Vectors map[string][]float32
	Size    int
}

// LoadWord2VecModel loads a Word2Vec model from a file
// and returns a Word2VecModel struct
// Doing this by hand is probably a mistake and may be done better with a library
func LoadWord2VecModel(modelPath string) (*Word2VecModel, error) {
	file, err := os.Open(modelPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := bufio.NewReader(file)

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

// GetVectorEmbedding returns the vector embedding of a token if it exists in the model
func GetVectorEmbedding(token string, model *Word2VecModel) []float32 {
	vec, ok := model.Vectors[token]
	if !ok {
		return make([]float32, model.Size)
	}
	return vec
}
