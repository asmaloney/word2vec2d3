package main

// Processes a word2vec binary file using t-SNE to produce a CSV which d3.js can visualize
// using a scatterplot.

import (
	"fmt"
	"os"

	"github.com/danaugrs/go-tsne/tsne"
	"gonum.org/v1/gonum/mat"

	"github.com/asmaloney/word2vec2d3/W2VBin"
)

func main() {
	// binary input file is produced by word2vec
	inputFile := "../word2vec/data/vectors.bin"
	// maximum number of words to use from the file
	maxWords := 1250

	// t-SNE paramaters
	perplexity := 5.0
	learningRate := 300.0
	iterations := 1500

	// output file name
	outputFile := "data.csv"

	// Load word2vec binary data
	words, vectors, err := W2VBin.Load(inputFile, maxWords)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Process using t-SNE
	fmt.Printf("t-SNE:\n\tperplexity = %f\n\tlearning rate = %f\n\titerations = %d\n",
		perplexity, learningRate, iterations)
	t := tsne.NewTSNE(2, perplexity, learningRate, iterations, true)
	embedding := t.EmbedData(vectors, func(iter int, divergence float64, embedding mat.Matrix) bool {
		if iter%10 == 0 {
			fmt.Printf(" [%d]: divergence = %v\n", iter, divergence)
		}
		return false
	})

	// Write it out as CSV
	err = writeDataCSV(outputFile, words, embedding)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func writeDataCSV(file string, words []string, embedding mat.Matrix) (err error) {
	f, err := os.Create(file)
	if err != nil {
		return
	}
	defer f.Close()

	fmt.Fprintf(f, "word,x,y\n")

	for i, word := range words {
		fmt.Fprintf(f, "%s,%v,%v\n", word, embedding.At(i, 0), embedding.At(i, 1))
	}

	return nil
}
