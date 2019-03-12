package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"os"
	"path/filepath"
)

// GrayPixel is a data structure to represent the gray and alpha value of a pixel.
type GrayPixel struct {
	y uint8
	a uint8
}

func main() {
	// define command line flags
	blurFlagPtr := flag.Bool("blur", true, "perform gaussian blur before edge detection (optional, default: true)")
	inputFileArgPtr := flag.String("input", "", "path to input file (required)")
	outputFileArgPtr := flag.String("output", "out.jpg", "path to output file (optional, default: out.jpg")
	minThresholdArgPtr := flag.Float64("min", float64(0.2), "ratio of lower threshold (optional, default: 0.2")
	maxThresholdArgPtr := flag.Float64("max", float64(0.6), "ratio of upper threshold (optional, default: 0.6")
	// parse command line flags and arguments
	flag.Parse()
	// check for required arguments, exit if empty path is provided
	if *inputFileArgPtr == "" {	// if no input filepath was specified, print message and exit
		fmt.Println("No path to input file specified, nothing to do.")
		return
	}
	// check threshold ratio arguments, exit if invalid values are given
	if !isValidRatioValue(*minThresholdArgPtr) || !isValidRatioValue(*maxThresholdArgPtr) {
		fmt.Println("Invalid value for threshold ratio given, exiting.")
		return
	}

	// register the jpeg and png formats with the image library
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

	// open the image specified by input argument
	pixels := openImage(*inputFileArgPtr)
	// perform Canny edge detection on the pixel array
	pixels = CannyEdgeDetect(pixels, *blurFlagPtr, *minThresholdArgPtr, *maxThresholdArgPtr)
	// write result to image file
	writeImage(pixels, *outputFileArgPtr)

}

// openImage opens the image given by a path string, converts it to grayscale and returns the pixels as a
// two-dimensional array
func openImage(path string) [][]GrayPixel {
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close() // opened for reading, no error checking needed

	// read the image data and convert to array of GrayPixel objects
	pixels, err := getPixelArray(file)
	if err != nil {
		log.Fatal(err)
	}

	return pixels
}

// writeImage takes a two-dimensional array of grayvalue pixels and writes the image to disc. The format of the image
// is specified by the path string. Suppoerted formats are png and jpg. If the path string is not detected as png a jpg
// is written by default.
func writeImage(pixels [][]GrayPixel, path string) {
	// create grayscale image from the pixel array and write it to disk
	grayImg := getImageFromArray(pixels)
	outFile, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}
	// determine what image file type it should be
	ext := filepath.Ext(path)
	if ext == "png" {
		err = png.Encode(outFile, grayImg)
	} else {
		opts := jpeg.Options{95}
		err = jpeg.Encode(outFile, grayImg, &opts)
	}
	if err != nil {
		log.Fatal(err)
	}
}

// getPixelArray reads the given file as an image and returns a two-dimensional array of GrayPixel objects. The values
// in the returned array are stored in the way that arr[m][n] refers to the n-th column of the m-th row of the image
// data.
func getPixelArray(file io.Reader) ([][]GrayPixel, error) {
	var pixelArr [][]GrayPixel

	// load the image from given file and determine image bounds
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	height := img.Bounds().Max.Y
	width := img.Bounds().Max.X

	// build pixel array row by row from image data
	for y := 0; y < height; y++ {
		var row []GrayPixel
		for x := 0; x < width; x++ {
			pixel := img.At(x, y)
			grayPixel := rgbaToGrayPixel(pixel)
			row = append(row, grayPixel)
		}
		pixelArr = append(pixelArr, row)
	}

	return pixelArr, nil
}

// getImageFromArray takes pixel information from the given two-dimensional array and creates a corresponding image.
func getImageFromArray(pixels [][]GrayPixel) *image.Gray {
	// construct bounding rectangle and create clear grayscale image
	bounds := image.Rect(0, 0, len(pixels[0]), len(pixels))
	img := image.NewGray(bounds)

	// set pixel values
	for y := 0; y < len(pixels); y++ {
		for x := 0; x < len(pixels[y]); x++ {
			img.SetGray(x, y, color.Gray{pixels[y][x].y})
		}
	}

	return img
}

// isValidRatioValue checks whether the given value lies between 0.0 and 1.0 thus providing a valid value for a ratio.
func isValidRatioValue(x float64) bool {
	if (x >= float64(0)) && (x <= float64(1)) {
		return true
	}
	return false
}

// rgbaToGrayPixel converts the given Color object to a GrayPixel object.
func rgbaToGrayPixel(pixel color.Color) GrayPixel {
	_, _, _, a := pixel.RGBA()
	gray := color.GrayModel.Convert(pixel).(color.Gray).Y

	return GrayPixel{gray, uint8(a >> 8)}
}