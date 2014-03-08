package main

import (
	"fmt"
	"github.com/skelterjohn/go.matrix" // daa59528eefd43623a4c8e36373a86f9eef870a2
	"github.com/youpy/go-wav"
	"io"
	"log"
	"os"
)

func main() {
	// I have no idea what I am doing
	fmt.Println("Bop")
	if len(os.Args) <= 1 {
		log.Fatal("You need to tell me what file to encode.")
	}
	file, _ := os.Open(os.Args[1])

	output, _ := os.OpenFile("./out.wav", os.O_CREATE, 600)

	reader := wav.NewReader(file)

	// Settings for output
	var numSamples uint32 = 999999
	var numChannels uint16 = 2
	var sampleRate uint32 = 44100
	var bitsPerSample uint16 = 16
	writer := wav.NewWriter(output, numSamples, numChannels, sampleRate, bitsPerSample)
	BlockesProcessed := 0
	// End of settings for output
	var SampleBlock []float64
	SampleBlock = make([]float64, 8)
	for {
		samplez := make([]wav.Sample, 0)
		samples, err := reader.ReadSamples()
		// Samples will return (usally) 2048 samples in a single sitting

		log.Println(BlockesProcessed)
		if err == io.EOF {
			break
		}

		BlockesProcessed++
		var SamplePointer int = 0 // max is 8
		for _, sample := range samples {
			sam := wav.Sample{}
			L := reader.IntValue(sample, 0)
			R := reader.IntValue(sample, 1)
			MonoVal := (L + R) / 2

			// Now we add it to the staging array
			SampleBlock[SamplePointer] = float64(MonoVal)
			SamplePointer++
			if SamplePointer == 8 {
				GetPolyResults([]float64{0, 1, 2, 3, 4, 5, 6, 7}, SampleBlock)
				SamplePointer = 0
			}
			sam.Values[0] = L
			sam.Values[0] = R
			samplez = append(samplez, sam)
		}

		// fmt.Println(len(samplez), len(samples))
		writer.WriteSamples(samplez)
	}

}

var degree = 5

func GetPolyResults(xGiven []float64, yGiven []float64) []float64 {
	m := len(yGiven)
	if m != len(xGiven) {
		return []float64{0, 0, 0, 0, 0} // Send it back, There is nothing sane here.
	}
	if m < 5 {
		// Prevent the processing of really small datasets, This is becauase there
		// appears to be a bug in the libary that will trigger a crash in the go.matrix
		// if some (small) amount of values are entered. I don't know why this happens
		// (Otherwise I would have fixed it) but the URL for the github issue is:
		// https://github.com/skelterjohn/go.matrix/issues/11
		return []float64{0, 0, 0, 0, 0} // Send it back, There is nothing sane here.
	}
	n := degree + 1
	y := matrix.MakeDenseMatrix(yGiven, m, 1)
	x := matrix.Zeros(m, n)
	for i := 0; i < m; i++ {
		ip := float64(1)
		for j := 0; j < n; j++ {
			x.Set(i, j, ip)
			ip *= xGiven[i]
		}
	}

	q, r := x.QR()
	qty, err := q.Transpose().Times(y)
	if err != nil {
		log.Println(err)
		return []float64{0, 0, 0, 0, 0} // Send it back, There is nothing sane here.
	}
	c := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		c[i] = qty.Get(i, 0)
		for j := i + 1; j < n; j++ {
			c[i] -= c[j] * r.Get(i, j)
		}
		c[i] /= r.Get(i, i)
	}
	// log.Println(c)
	return c
}
