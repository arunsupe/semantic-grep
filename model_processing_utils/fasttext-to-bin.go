/*
A small utility to convert FastText models to Word2Vec format.
The input file should be a FastText model in text format.
The output file will be a Word2Vec binary model.

Usage:
  fasttext-to-bin -input <input_fasttext_file> -output <output_word2vec_file>

Example:
  fasttext-to-bin -input model.bin -output model.bin

Or stream from stdin:
  curl -s 'https://dl.fbaipublicfiles.com/fasttext/vectors-crawl/cc.fr.300.vec.gz' \
  | gunzip -c \
  | fasttext-to-bin -input - -output models/fasttext/cc.fr.300.bin
*/

package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func convertFastTextToWord2Vec(input io.Reader, outputFile string) error {
	// Open output file
	out, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}
	defer out.Close()

	writer := bufio.NewWriter(out)
	defer writer.Flush()

	scanner := bufio.NewScanner(input)

	// Read header
	if !scanner.Scan() {
		return fmt.Errorf("error reading header: %v", scanner.Err())
	}
	header := strings.Fields(scanner.Text())
	if len(header) != 2 {
		return fmt.Errorf("invalid header format")
	}

	vocabSize, err := strconv.Atoi(header[0])
	if err != nil {
		return fmt.Errorf("invalid vocabulary size: %v", err)
	}

	vectorSize, err := strconv.Atoi(header[1])
	if err != nil {
		return fmt.Errorf("invalid vector size: %v", err)
	}

	// Write text header
	if _, err := fmt.Fprintf(writer, "%d %d\n", vocabSize, vectorSize); err != nil {
		return fmt.Errorf("error writing text header: %v", err)
	}

	// Process each line
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != vectorSize+1 {
			return fmt.Errorf("invalid line format: expected %d fields, got %d", vectorSize+1, len(parts))
		}

		word := parts[0]
		if _, err := writer.WriteString(word); err != nil {
			return fmt.Errorf("error writing word: %v", err)
		}
		if err := writer.WriteByte(' '); err != nil {
			return fmt.Errorf("error writing space: %v", err)
		}

		vector := make([]float32, vectorSize)
		for i := 0; i < vectorSize; i++ {
			value, err := strconv.ParseFloat(parts[i+1], 32)
			if err != nil {
				return fmt.Errorf("error parsing float: %v", err)
			}
			vector[i] = float32(value)
		}

		if err := binary.Write(writer, binary.LittleEndian, vector); err != nil {
			return fmt.Errorf("error writing vector: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning input: %v", err)
	}

	return nil
}

func main() {
	// Define command-line flags
	inputFileFlag := flag.String("input", "", "Input FastText file (use '-' for stdin)")
	outputFileFlag := flag.String("output", "", "Output Word2Vec file. End in .bin")
	flag.Parse()

	// Validate flags
	if *inputFileFlag == "" || *outputFileFlag == "" {
		fmt.Println("Usage: fasttext-to-bin -input <input_fasttext_file> -output <output_word2vec_file>")
		os.Exit(1)
	}

	var input io.Reader

	// Check if input is from stdin or a file
	if *inputFileFlag == "-" {
		input = os.Stdin
	} else {
		file, err := os.Open(*inputFileFlag)
		if err != nil {
			fmt.Printf("Error opening input file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		input = file
	}

	// Convert FastText to Word2Vec
	err := convertFastTextToWord2Vec(input, *outputFileFlag)
	if err != nil {
		fmt.Printf("Error during conversion: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Conversion complete. Word2Vec binary model saved as %s\n", *outputFileFlag)
}
