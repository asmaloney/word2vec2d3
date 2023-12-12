// Package W2VBin provides a function to load word2vec binary files and return the data
// as a slice of words and a matrix of their vectors.
package W2VBin

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"

	"gonum.org/v1/gonum/mat"
)

type data struct {
	bytes []byte

	current int

	eof bool
}

func (d *data) atEOF() bool {
	if d.current >= len(d.bytes) {
		d.eof = true
	}

	return d.eof
}

func (d *data) nextLine() (err error) {
	if d.bytes[d.current] != '\n' {
		return fmt.Errorf("Expected a newline, found %d @ %d", d.bytes[d.current], d.current)
	}

	d.current++
	return nil
}

func (d *data) nextString() string {
	if d.atEOF() {
		return ""
	}

	end := 0

	for {
		end++

		pos := d.current + end

		if pos >= len(d.bytes) {
			d.eof = true
			return ""
		}

		char := d.bytes[pos]

		if (char == ' ') || (char == '\n') {
			str := string(d.bytes[d.current:pos])
			d.current += end + 1
			return str
		}
	}
}

func (d *data) nextFloat() float32 {
	if d.atEOF() {
		return 0.0
	}

	val := d.bytes[d.current : d.current+4]

	d.current += 4

	bits := binary.LittleEndian.Uint32(val)
	return math.Float32frombits(bits)
}

// Load will load "file" and return the data as a slice of words and a matrix of their vectors.
// Setting "wordLimit" to 0 will return all the data. Setting it to a positive number will return
// data for the first "wordLimit" words in the data set.
func Load(file string, wordLimit int) (words []string, vectors mat.Matrix, err error) {
	fileBytes, err := os.ReadFile(file)
	if err != nil {
		return
	}

	d := data{bytes: fileBytes, current: 0, eof: false}

	numWordsStr := d.nextString()
	if d.eof {
		return
	}

	numWords, err := strconv.Atoi(numWordsStr)
	if err != nil {
		return
	}

	numDimensionsStr := d.nextString()
	if d.eof {
		return
	}

	numDimensions, err := strconv.Atoi(numDimensionsStr)
	if err != nil {
		return
	}

	fmt.Printf("%q : %d bytes\n", file, len(fileBytes))
	fmt.Printf("\t%d words; %d dimensions\n", numWords, numDimensions)

	if wordLimit != 0 {
		numWords = min(numWords, wordLimit)
		fmt.Printf("\t(limiting to %d words)\n", numWords)
	}

	words = make([]string, numWords-1)
	vectorData := make([]float64, (numWords-1)*numDimensions)

	count := 0
	for i := 0; i < numWords; i++ {
		word := d.nextString()
		if d.eof {
			break
		}

		// we skip the "</s>" "word" which was used to indicate newlines in the data
		isNewline := strings.Compare(word, "</s>") == 0

		if !isNewline {
			words[count] = word
			// fmt.Printf("[%d] %s\n", count, words[count])
		}

		for dim := 0; dim < numDimensions; dim++ {
			val := d.nextFloat()
			if d.eof {
				break
			}

			if !isNewline {
				vectorData[count*numDimensions+dim] = float64(val)
			}
		}

		err = d.nextLine()
		if err != nil {
			return
		}

		if !isNewline {
			count++
		}
	}

	return words, mat.NewDense(count, numDimensions, vectorData), nil
}
