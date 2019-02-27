package main

import (
	"errors"
	"fmt"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/combin"
	"math"
)

type direction int

const (
	HORIZONTAL direction = iota
	VERTICAL
)

var SOBEL_X = []float64{1, 0, -1, 2, 0, -2, 1, 0, -1} // matrix values for sobel filter (x-component)
var SOBEL_Y = []float64{1, 2, 1, 0, 0, 0, -1, -2, -1} // matrix values for sobel filter (y-component)

func CannyEdgeDetect(pixels [][]GrayPixel) [][]GrayPixel {
	pixels = gaussianBlur(pixels, 5)
	pixels, angles := sobel(pixels)
	fmt.Printf("angles:\n")
	for y:=0; y<len(angles); y++ {
		fmt.Printf("row %d:", y)
		for x:=0; x<len(angles[0]); x++ {
			fmt.Printf("col %d: %f ", x, angles[y][x])
		}
		fmt.Printf("\n")
	}

	return pixels
}

// sobel performs the sobel edge detection filter method on the given image. In addition it returns the gradient
// directions of all pixels as a two-dimensional array of degree values.
func sobel(pixels [][]GrayPixel) ([][]GrayPixel, [][]float64){
	var result [][]GrayPixel
	var directions [][]float64
	// build sobel filter kernels
	sobel_X := *mat.NewDense(3, 3, SOBEL_X)
	sobel_Y := *mat.NewDense(3, 3, SOBEL_Y)
	// apply the two kernels to all pixels
	for y:=0; y<len(pixels); y++ {
		var resultRow []GrayPixel
		var angleRow []float64
		for x:=0; x<len(pixels[y]); x++ {
			var angle float64
			// get matrices with sorrounding pixel values
			imagePane := getSorroundingPixelMatrix(pixels, y, x, 3)
			// convolve with kernel for x and y direction
			sobelRes_X := convolve(imagePane, sobel_X)
			sobelRes_Y := convolve(imagePane, sobel_Y)
			// combine results
			combinedRes := uint8(math.Sqrt(math.Pow(sobelRes_X, 2) + math.Pow(sobelRes_Y, 2)))
			resultRow = append(resultRow, GrayPixel{combinedRes, uint8(255)})
			// calculate gradient direction
			if (sobelRes_X == float64(0)) || (sobelRes_Y == float64(0)) {
				angle = float64(0)
			} else {
				angle = math.Atan(sobelRes_Y / sobelRes_X)
			}
			angle = angle * (180/math.Pi)	// convert from radians to degree
			angleRow = append(angleRow, angle)
		}
		result = append(result, resultRow)
		directions = append(directions, angleRow)
	}

	return result, directions
}

// gaussianBlur performs a gaussian blur filtering on the given image by using a kernel of the given size. Note that the
// kernel size must be odd, otherwise the function will panic. The blurred image is returned.
func gaussianBlur(pixels [][]GrayPixel, kernelSize uint) [][]GrayPixel {
	if kernelSize%2 == 0 { // we only allow odd kernel sizes, panic if it is even
		panic(errors.New("size of kernel must be odd"))
	}
	var result [][]GrayPixel
	kernel := getPascalTriangleRow(kernelSize - 1) // to get n kernel elements we need the (n-1)th row
	kernel = normalizeVec(kernel)                  // normalize kernel so we don't change brightness of the pixels
	// iterate over each pixel of the image and apply the gaussian kernel
	for y := 0; y < len(pixels); y++ {
		var resultRow []GrayPixel
		for x := 0; x < len(pixels[y]); x++ {
			vecVert := getPixelVector(pixels, y, x, kernel.Len(), VERTICAL)
			vecHor := getPixelVector(pixels, y, x, kernel.Len(), HORIZONTAL)
			verticalSum := innerProduct(vecVert, kernel)
			horizontalSum := innerProduct(vecHor, kernel)
			combinedRes := uint8(math.Sqrt(verticalSum*verticalSum + horizontalSum*horizontalSum))	// combine both sums
			resultRow = append(resultRow, GrayPixel{combinedRes, 255})
		}
		result = append(result, resultRow)
	}

	return result
}

// getSorroundingPixelMatrix returns a matrix that contains the pixels sorrounding the pixel at the given location. The
// resulting matrix is a square with the width defined by the length parameter and is centered at the given pixel
// location. Note that this function panics if the given length is an even number.
func getSorroundingPixelMatrix(pixels [][]GrayPixel, posY, posX int, length int) mat.Dense {
	if length%2 == 0 { // length must be an odd number
		panic(errors.New("length must be odd number"))
	}

	var values []float64 // return values
	var currentPixel GrayPixel
	padding := (length / 2) // how much pixels to left, right, top and bottom we need
	// get limits for loop indices
	minX := posX - padding
	minY := posY - padding
	maxX := posX + padding
	maxY := posY + padding
	height := len(pixels)
	width := len(pixels[0])

	var curY, curX int
	for y:=minY; y<=maxY; y++ {
		if y<0 {	// top border pixels
			curY = posY + abs(y)
		} else if y >= height {	// bottom border pixels
			overlap := y - height + 1 // add 1 because array length is bigger than last valid index
			curY = posY-overlap
		} else {
			curY = y
		}
		for x:=minX; x<=maxX; x++ {
			if x<0 {	// left border pixels
				curX = posX + abs(x)
			} else if x>=width {	// right border pixels
				overlap := x - width + 1 // add 1 because array length is bigger than last valid index
				curX = posX-overlap
			} else {
				curX = x
			}
			// append pixel value
			currentPixel = pixels[curY][curX]
			values = append(values, float64(currentPixel.y))
		}
	}

	return *mat.NewDense(length, length, values)
}

// getPixelVector returns a vector of given length from the given [][]GrayPixel. The pixels are taken from the
// position given by x and y and from the nearby area as denoted by the direction parameter. In case of border pixels
// pixel values mirrored from inside the image are used instead. The fact that an equal amount of pixels is to be
// returned from the left and right side of the given position requires the length parameter to be an odd number. In
// cases of length being an even number the function panics.
func getPixelVector(pixels [][]GrayPixel, posY, posX int, length int, dir direction) mat.VecDense {
	if length%2 == 0 { // length must be an odd number
		panic(errors.New("length must be odd number"))
	}

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

// innerProduct calculates the inner product of the two given vectors. This means that the result is the sum of the
// products of the first elements of both vectors and the sum of the second elements of both vectors and so on. Note
// that this function panics if the length of both given vectors is not equal.
func innerProduct(pixels, kernel mat.VecDense) float64 {
	if pixels.Len() != kernel.Len() { // vectors must have equal length
		panic(errors.New("length of given vectors must be equal"))
	}

	var result float64 = 0
	for i := 0; i < pixels.Len(); i++ {
		result += pixels.At(i, 0) * kernel.At(i, 0)
	}

	return result
}

// convolve returns the result of the convolution operation with the two given matrices. Note that this function will
// panic if the dimensions of the matrices are not identical.
func convolve(m1, m2 mat.Dense) float64 {
	row_1, col_1 := m1.Dims()
	row_2, col_2 := m2.Dims()
	if row_1 != row_2 || col_1 != col_2 {
		panic(errors.New("invalid matrix dimensions for convolution operation"))
	}

	var result float64 = 0
	rows, cols := m1.Dims()

	for y:=0; y<rows; y++ {
		for x:=0; x<cols; x++ {
			result += m1.At(y, x) * m2.At(y, x)
		}
	}

	return result
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
	var sum float64 = 0
	for i := 0; i < v.Len(); i++ {
		sum += v.At(i, 0)
	}
	// create result vector that is given vector divided by sum
	var result mat.VecDense
	result.ScaleVec(1/sum, v.SliceVec(0, v.Len()))
	return result
}

// abs returns the absolute value of the given int.
func abs(x int) int {
	if x < 0 {
		return (-x)
	} else {
		return x
	}
}
