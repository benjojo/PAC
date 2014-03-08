package main

import (
	"flag"
	"fmt"
	"github.com/skelterjohn/go.matrix" // daa59528eefd43623a4c8e36373a86f9eef870a2
	"github.com/youpy/go-wav"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

var EncodeBlockSize int = 8
var LastBlockSample int = 0
var PolySize int = 5

func main() {
	// I have no idea what I am doing
	Encoding := flag.Bool("encode", false, "true if you want to encode")
	Decoding := flag.Bool("decode", false, "true if you want to decode")
	Filename := flag.String("in", "", "What file the output sould be read from")
	OutputFile := flag.String("out", "", "What file the output sould be written to")
	flag.Parse()

	if *Encoding && *Decoding {
		log.Fatal("You can't do both!")
	}
	if *Filename == "" || *OutputFile == "" {
		log.Fatal("Please give both input and output.")
	}

	fmt.Println("Bop")
	if len(os.Args) <= 1 {
		log.Fatal("You need to tell me what file to encode.")
	}
	if *Encoding {
		Encode(*Filename, *OutputFile)
	} else {
		Decode(*Filename, *OutputFile)
	}

}

func Encode(filename, OutputFile string) {
	log.Println("Encoding file...")

	file, _ := os.Open(filename)

	outputpac, _ := os.OpenFile(OutputFile, os.O_CREATE, 600)

	reader := wav.NewReader(file)

	BlockesProcessed := 0
	// End of settings for output
	var SampleBlock []float64
	SampleBlock = make([]float64, EncodeBlockSize)
	for {
		samplez := make([]wav.Sample, 0)
		samples, err := reader.ReadSamples()
		// Samples will return (usally) 2048 samples in a single sitting

		log.Println(BlockesProcessed)
		if err == io.EOF {
			break
		}

		BlockesProcessed++
		XBlock := make([]float64, EncodeBlockSize)
		for k, _ := range XBlock {
			XBlock[k] = float64(k)
		}
		var SamplePointer int = 0 // max is EncodeBlockSize
		for _, sample := range samples {
			sam := wav.Sample{}
			L := reader.IntValue(sample, 0)
			R := reader.IntValue(sample, 1)
			MonoVal := (L + R) / 2

			// Now we add it to the staging array
			SampleBlock[SamplePointer] = float64(MonoVal)
			SamplePointer++
			if SamplePointer == EncodeBlockSize {
				out := GetPolyResults(XBlock, SampleBlock)
				Line := fmt.Sprintf("%f,%f,%f,%f,%f\n", out[0], out[1], out[2], out[3], out[4])
				outputpac.Write([]byte(Line))
				SamplePointer = 0
			}
			sam.Values[0] = L
			sam.Values[0] = R
			samplez = append(samplez, sam)
		}

		// fmt.Println(len(samplez), len(samples))
	}

}

func Decode(filename, OutputFile string) {
	log.Println("Decoding file")
	b, e := ioutil.ReadFile(filename)
	if e != nil {
		log.Fatal("cannot read input file.")
	}
	outputwav, e := os.OpenFile(OutputFile, os.O_CREATE, 600)
	if e != nil {
		log.Fatal("cannot open output file")
	}

	var numSamples uint32 = 999999
	var numChannels uint16 = 2
	var sampleRate uint32 = 44100
	var bitsPerSample uint16 = 16
	writer := wav.NewWriter(outputwav, numSamples, numChannels, sampleRate, bitsPerSample)

	lines := strings.Split(string(b), "\n")
	samplez := make([]wav.Sample, 0)

	for _, line := range lines {
		var prams []float64
		prams = make([]float64, 5)
		bits := strings.Split(line, ",")

		if len(bits) != 5 {
			if len(bits) == 1 {
				break
			}
			log.Fatal("Invalid PAC file. ", len(bits))
		}
		for k, v := range bits {
			prams[k], e = strconv.ParseFloat(v, 64)
			if e != nil {
				log.Fatal("unable to decode part of PAC file")
			}
		}
		out := GetSamplesFromPoly(prams)

		for _, v := range out {
			sam := wav.Sample{}

			sam.Values[0] = v
			sam.Values[1] = v
			samplez = append(samplez, sam)
		}
	}
	writer.WriteSamples(samplez)
}

func GetSamplesFromPoly(prams []float64) (out []int) {
	out = make([]int, EncodeBlockSize)
	for k, _ := range out {
		out[k] = int(
			(prams[4] * math.Pow(float64(k), 4)) +
				(prams[3] * math.Pow(float64(k), 3)) +
				(prams[2] * math.Pow(float64(k), 2)) +
				(prams[1] * float64(k)) + prams[0])
		if k == 0 {
			out[k] = (out[k] + LastBlockSample) / 2
		}
	}
	LastBlockSample = out[EncodeBlockSize-1]
	return out
}

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
	n := PolySize + 1
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
