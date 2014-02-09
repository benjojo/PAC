package main

import (
	"fmt"
	"github.com/youpy/go-wav"
	"io"
	"io/ioutil"
	"os"
)

func main() {
	// I have no idea what I am doing
	fmt.Println("Bop")
	file, _ := os.Open("./test.wav")
	output, _ := os.Open("./out.wav")
	reader := wav.NewReader(file)

	// Settings for output
	var numSamples uint32 = 2
	var numChannels uint16 = 2
	var sampleRate uint32 = 44100
	var bitsPerSample uint16 = 16
	// End of settings for output
	for {
		samples, err := reader.ReadSamples()
		fmt.Println(len(samples))
		if err == io.EOF {
			break
		}

		// for _, sample := range samples {
		// fmt.Printf("L/R: %d/%d\n", reader.IntValue(sample, 0), reader.IntValue(sample, 1))
		// }
	}
}
