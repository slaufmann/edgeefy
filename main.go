package main

import (
	"fmt"
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
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	file, err := os.Open("./logo.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	pixels, err := getPixelArray(file)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Println(pixels)
}

// getPixelArray reads the given file as an image and returns a two-dimensional array of GrayPixel objects.
func getPixelArray(file io.Reader) ([][]GrayPixel, error) {
	var pixelArr [][]GrayPixel

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	height := img.Bounds().Max.Y
	width := img.Bounds().Max.X

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

// rgbaToGrayPixel converts the given color object to a GrayPixel object.
func rgbaToGrayPixel(pixel color.Color) GrayPixel {
	_, _, _, a := pixel.RGBA()
	gray := color.GrayModel.Convert(pixel).(color.Gray).Y

	return GrayPixel{gray, uint8(a >> 8)}
}
