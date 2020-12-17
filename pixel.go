package main

type Pixel struct {
	R int
	G int
	B int
	A int
}

func getRed(lePixel Pixel) int {
	return lePixel.R
}

func getGreen(lePixel Pixel) int {
	return lePixel.G
}

func getBlue(lePixel Pixel) int {
	return lePixel.B
}

func getAlpha(lePixel Pixel) int {
	return lePixel.A
}
