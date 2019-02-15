package main

import (
	"errors"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/combin"
)

type direction int

const (
	HORIZONTAL direction = iota
	VERTICAL
)

func CannyEdgeDetect(pixels GrayPixelImage) GrayPixelImage {
	pixels = gaussianBlur(pixels, 7)

	return pixels
}

// gaussianBlur performs a gaussian blur filtering on the given image by using a kernel of the given size. Note that the
// kernel size must be odd, otherwise the function will panic. The blurred image is returned.
func gaussianBlur(pixels GrayPixelImage, kernelSize uint) GrayPixelImage {
	if kernelSize%2 == 0 { // we only allow odd kernel sizes, panic if it is even
		panic(errors.New("size of kernel must be odd"))
	}
	kernel := getPascalTriangleRow(kernelSize - 1) // to get n kernel elements we need the (n-1)th row
	kernel = normalizeVec(kernel)                  // normalize kernel so we don't change brightness of the pixels
	applyKernel(pixels, kernel)                    // apply the kernel to blur the image

	return pixels
}

// getPascalTriangleRow returns the row of a pascal triangle with the given index in the form of a dense column vector.
func getPascalTriangleRow(index uint) mat.VecDense {
	size := int(index + 1)          // we need an array that is 1 bigger than the index of the requested row
	values := make([]float64, size) // array to store row values
	// calculate the row values via the binomial coefficient
	for i := 0; i < size; i++ {
		values[i] = float64(combin.Binomial(int(index), i))
	}
	// return row as dense vector
	result := mat.NewVecDense(size, values)
	return *result
}

// normalizeVec normalizes a given vector by summing up the elements and returning a new vector with an element sum of 1.
func normalizeVec(v mat.VecDense) mat.VecDense {
	// calculate the sum of all vector elements
	sum := float64(0)
	for i := 0; i < v.Len(); i++ {
		sum += v.At(i, 0)
	}
	// create result vector that is given vector divided by sum
	var result mat.VecDense
	result.ScaleVec(1/sum, v.SliceVec(0, v.Len()))
	return result
}

// applyKernel the given kernel to the provided image. The border option used is mirroring of image
// pixels.
func applyKernel(pixels GrayPixelImage, kernel mat.VecDense) GrayPixelImage {
	// iterate over each pixel of the given image and apply the kernel
	for y := 0; y < len(pixels); y++ {
		for x := 0; x < len(pixels[y]); x++ {
			// construct vector with vertical pixel values
			// call mulAndSum with kernel
			verticalSum := mulAndSum(getPixelVector(pixels, y, x, kernel.Len(), VERTICAL), kernel)
			// construct vector with horizontal pixel values
			// call mulAndSum again
			horizontalSum := mulAndSum(getPixelVector(pixels, y, x, kernel.Len(), HORIZONTAL), kernel)
			// assign new value to pixel
			pixels[y][x].y = uint8(verticalSum + horizontalSum)
		}
	}
	return pixels
}

// mulAndSum calculates the inner product of the two given vectors. This means that the result is the sum of the
// products of the first element of both vectors and the second element of both vectors and so on. Note that this
// function panics if the length of both given vectors is not equal.
func mulAndSum(pixels mat.VecDense, kernel mat.VecDense) float64 {
	if pixels.Len() != kernel.Len() { // vectors must have equal length
		panic(errors.New("length of given vectors must be equal"))
	}

	result := float64(0)

	for i := 0; i < pixels.Len(); i++ {
		result += pixels.At(i, 0) * kernel.At(i, 0)
	}

	return result
}

// getPixelVector returns a vector of given length from the given GrayPixelImage. The pixels are taken from the
// position given by x and y and from the nearby area as denoted by the direction parameter.
func getPixelVector(pixels GrayPixelImage, posY int, posX int, length int, dir direction) mat.VecDense {
	var values []float64 // return values
	var currentPixel GrayPixel
	padding := (length / 2) // how much pixels to either the left and right or top and bottom we need

	switch dir {
	case HORIZONTAL:
		minX := posX - padding
		maxX := posX + padding
		for i := minX; i <= maxX; i++ {
			rowLength := len(pixels[posY])
			if i < 0 { // left border pixels
				currentPixel = pixels[posY][posX+abs(i)]
			} else if i >= rowLength { // right border pixels
				overlap := i - rowLength + 1 // add 1 because array length is bigger than last valid index
				currentPixel = pixels[posY][posX-overlap]
			} else { // non-border pixels
				currentPixel = pixels[posY][i]
			}
			values = append(values, float64(currentPixel.y))

		}
	case VERTICAL:
		minY := posY - padding
		maxY := posY + padding
		for i := minY; i <= maxY; i++ {
			columnLength := len(pixels)
			if i < 0 { // top border pixels
				currentPixel = pixels[posY+abs(i)][posX]
			} else if i >= columnLength { // bottom border pixels
				overlap := i - columnLength + 1 // add 1 because array length is bigger than last valid index
				currentPixel = pixels[posY-overlap][posX]
			} else { // non-border pixels
				currentPixel = pixels[i][posX]
			}
			values = append(values, float64(currentPixel.y))
		}
	}

	return *mat.NewVecDense(len(values), values)
}

// abs returns the absolute value of the given int.
func abs(x int) int {
	if x < 0 {
		return (-x)
	} else {
		return x
	}
}
