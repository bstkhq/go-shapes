package main

import (
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
)

func Usage() {
	fmt.Printf("Usage:\n\tgo run internal/gen/sigma_kernels.go binomial\n")
	fmt.Printf("\tgo run internal/gen/sigma_kernels.go squareDiv 2.0\n")
	fmt.Printf("\tgo run internal/gen/sigma_kernels.go sigmaDiv 3.0\n")
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 || os.Args[1] != "binomial" && len(os.Args) != 3 {
		Usage()
	}

	genFunc := GaussianWeights
	squareDiv := false
	switch os.Args[1] {
	case "binomial":
		genFunc = BinomialWeights
	case "squareDiv":
		squareDiv = true
	case "sigmaDiv":
		// ok
	default:
		Usage()
	}

	var sigmaDiv float64 = 1.0
	if len(os.Args) == 3 {
		var err error
		sigmaDiv, err = strconv.ParseFloat(os.Args[2], 64)
		if err != nil {
			fmt.Printf("ERROR: %s", err)
			os.Exit(1)
		}
		if sigmaDiv <= 0 {
			fmt.Printf("div must be > 0")
			os.Exit(1)
		}
	}

	maxValues := 16
	fmt.Printf("var gaussKernels = [][%d]float32{\n", maxValues)
	for i := range maxValues - 1 {
		radius := i + 1
		sigma := float64(radius) / sigmaDiv // binomial-like
		if squareDiv {
			sigma = math.Sqrt(float64(radius) / sigmaDiv)
		}
		weights := genFunc(radius, sigma)
		m := slices.Max(weights)
		weights = weights[slices.Index(weights, m):]
		for len(weights) < maxValues {
			weights = append(weights, 0.0)
		}
		printWeightsRow(weights, maxValues, radius, "\t")
	}
	fmt.Printf("}\n")
}

func printWeightsRow(weights []float64, maxValues int, radius int, pre string) {
	fmt.Printf("%s{", pre)
	for k, w := range weights {
		if k == maxValues {
			fmt.Printf("}, // ...")
		}
		if k > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%.8f", w)
	}

	if len(weights) == maxValues {
		size := radius*2 + 1
		fmt.Printf("}, // %dx%d\n", size, size)
	} else {
		fmt.Print("\n")
	}
}

func GaussianWeights(radius int, sigma float64) []float64 {
	size := 2*radius + 1
	kernel := make([]float64, size)
	sum := 0.0

	twoSigmaSq := 2.0 * sigma * sigma
	for i := range size {
		x := float64(i - radius)
		weight := math.Exp(-(x * x) / twoSigmaSq)
		kernel[i] = weight
		sum += weight
	}

	for i := range kernel {
		kernel[i] /= sum
	}
	return kernel
}

func BinomialWeights(radius int, _ float64) []float64 {
	size := radius*2 + 1
	kernel := make([]float64, size)
	var sum float64
	for i, v := range pascalRow(size - 1) {
		kernel[i] = float64(v)
		sum += float64(v)
	}
	for i := range kernel {
		kernel[i] /= sum
	}
	return kernel
}

func pascalRow(n int) []int64 {
	row := make([]int64, n+1)
	row[0] = 1

	for k := 1; k <= n; k++ {
		row[k] = row[k-1] * int64(n-k+1) / int64(k)
	}
	return row
}
