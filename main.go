package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"os"
)

// GrayPixel is a data structure to represent the gray and alpha value of a pixel.
type GrayPixel struct {
	y uint8
	a uint8
}

func main() {
	// register the jpeg format with the image library and open the sample image
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	file, err := os.Open("./test2.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close() // opened for reading, no error checking needed

	// read the image data and convert to array of GrayPixel objects
	pixels, err := getPixelArray(file)
	if err != nil {
		log.Fatal(err)
	}

	// perform Canny edge detection on the pixel array
	pixels = CannyEdgeDetect(pixels, bool(false))

	// create grayscale image from the pixel array and write it to disk
	grayImg := getImageFromArray(pixels)
	outFile, err := os.Create("./out.jpg")
	if err != nil {
		log.Fatal(err)
	}
	opts := jpeg.Options{95}
	err = jpeg.Encode(outFile, grayImg, &opts)
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

// rgbaToGrayPixel converts the given Color object to a GrayPixel object.
func rgbaToGrayPixel(pixel color.Color) GrayPixel {
	_, _, _, a := pixel.RGBA()
	gray := color.GrayModel.Convert(pixel).(color.Gray).Y

	return GrayPixel{gray, uint8(a >> 8)}
}