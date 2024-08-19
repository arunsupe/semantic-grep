package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strings"
	"flag"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

type VecModel32bit struct {
	Vectors        map[string][]float32
	VectorsReduced map[string][]float32
	Size           int
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

// ReduceDimensions reduces the dimensions of the vectors to 100 using PCA
func (m *VecModel32bit) ReduceDimensions(targetDim int) error {
	vocabSize := len(m.Vectors)
	if vocabSize == 0 || m.Size <= 100 {
		return fmt.Errorf("no vectors to reduce or vector size is already 100 or less")
	}

	// Convert map to matrix
	data := make([]float64, 0, vocabSize*m.Size)
	words := make([]string, 0, vocabSize)
	for word, vector := range m.Vectors {
		words = append(words, word)
		for _, v := range vector {
			data = append(data, float64(v))
		}
	}

	originalMatrix := mat.NewDense(vocabSize, m.Size, data)

	// Perform PCA
	var pc stat.PC
	ok := pc.PrincipalComponents(originalMatrix, nil)
	if !ok {
		return fmt.Errorf("PCA computation failed")
	}

	// Get the principal component direction vectors
	var vec mat.Dense
	pc.VectorsTo(&vec)

	// Select the first targetDim columns of the principal components
	proj := mat.NewDense(vocabSize, targetDim, nil)
	proj.Mul(originalMatrix, vec.Slice(0, m.Size, 0, targetDim))

	// Convert reduced matrix back to map
	m.VectorsReduced = make(map[string][]float32, vocabSize)
	for i, word := range words {
		reducedVector := make([]float32, targetDim)
		for j := 0; j < targetDim; j++ {
			reducedVector[j] = float32(proj.At(i, j))
		}
		m.VectorsReduced[word] = reducedVector
	}

	return nil
}

// SaveReducedModel saves the reduced model to a file in the same binary format
func (m *VecModel32bit) SaveReducedModel(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Get the reduced vector size from the first vector in the map
	var reducedSize int
	for _, vector := range m.VectorsReduced {
		reducedSize = len(vector)
		break
	}

	// Write header: vocabSize and reduced vector size
	vocabSize := len(m.VectorsReduced)
	_, err = fmt.Fprintf(writer, "%d %d\n", vocabSize, reducedSize)
	if err != nil {
		return fmt.Errorf("failed to write header: %v", err)
	}

	// Write each word and its reduced vector
	for word, vector := range m.VectorsReduced {
		_, err := writer.WriteString(word + " ")
		if err != nil {
			return fmt.Errorf("failed to write word: %v", err)
		}

		for _, value := range vector {
			err := binary.Write(writer, binary.LittleEndian, value)
			if err != nil {
				return fmt.Errorf("failed to write vector value: %v", err)
			}
		}

		// Write a newline character after each vector
		err = writer.WriteByte('\n')
		if err != nil {
			return fmt.Errorf("failed to write newline: %v", err)
		}
	}

	// Flush the buffer to ensure all data is written to the file
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush writer: %v", err)
	}

	return nil
}

func main() {
	// Define command-line flags for input and output file paths
	inputFile := flag.String("input", "", "Path to the input Word2Vec model file")
	outputFile := flag.String("output", "", "Path to the output reduced model file")
	targetDim := flag.Int("dim", 100, "Target dimension for PCA reduction")
	flag.Parse()

	// Check if input and output paths are provided
	if *inputFile == "" || *outputFile == "" {
		fmt.Println("Please provide both input and output file paths using -input and -output flags.")
		return
	}

	// Load the model
	model := VecModel32bit{}
	err := model.LoadModel(*inputFile)
	if err != nil {
		fmt.Println("Error loading model:", err)
		return
	}

	// Reduce dimensions
	err = model.ReduceDimensions(*targetDim)
	if err != nil {
		fmt.Println("Error reducing dimensions:", err)
		return
	}

	// Save the reduced model
	err = model.SaveReducedModel(*outputFile)
	if err != nil {
		fmt.Println("Error saving reduced model:", err)
		return
	}

	fmt.Println("Reduced model saved successfully!")
}