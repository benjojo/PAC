package main

import (
	"fmt"
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
	var numSamples uint32 = 9999
	var numChannels uint16 = 2
	var sampleRate uint32 = 44100
	var bitsPerSample uint16 = 16
	writer := wav.NewWriter(output, numSamples, numChannels, sampleRate, bitsPerSample)
	a := 0
	// End of settings for output
	for {
		samplez := make([]wav.Sample, 0)
		samples, err := reader.ReadSamples()
		// Samples will return (usally) 2048 samples in a single sitting

		log.Println(len(samples))
		if err == io.EOF {
			break
		}

		a++
		for _, sample := range samples {
			sam := wav.Sample{}
			L := reader.IntValue(sample, 0)
			R := reader.IntValue(sample, 1)

			sam.Values[0] = L
			sam.Values[0] = R
			samplez = append(samplez, sam)

		}

		// fmt.Println(len(samplez), len(samples))
		writer.WriteSamples(samplez)
	}

}
