# edgeefy
Implementation of canny edge detection in Go

This projects implements edge detection in a given image using the so called [canny algorithm](https://en.wikipedia.org/wiki/Canny_edge_detector).
The implemented algorithm consists of the following steps:
1. perform gaussian blur (optional)
2. perform sobel filtering
3. apply non-maximum suppression
4. perform double thresholding
5. track edges by hysteresis

The implementation supports the input and output of jpg and png images.  

I started this project to get more familiar with the go programming language.
In the future I would like to use the edge detection functionality to transform images into something that looks like a grid representation of the main features of the image.
