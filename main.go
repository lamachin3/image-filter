package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
)

//type Pixel
type Pixel struct {
	R int
	G int
	B int
	A int
}

var filterMenu int
var inputFile, outputFile string
var height, width = 0, 0
var imgLoaded [][]Pixel

// crée une matrice à partir d'une image
// récupère les valeurs RGBA de chaque pixel
func getImg(file io.Reader) ([][]Pixel, error) {
	img, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	bounds := img.Bounds()
	width, height = bounds.Max.X, bounds.Max.Y

	for y := 0; y < height; y++ {
		var row []Pixel
		for x := 0; x < width; x++ {
			row = append(row, rgbaToPixel(img.At(x, y).RGBA()))
		}
		imgLoaded = append(imgLoaded, row)
	}
	return imgLoaded, nil
}

//conversion uint32 -> uint8
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257)}
}

func encode() {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(getRed(imgLoaded[y][x])),
				G: uint8(getGreen(imgLoaded[y][x])),
				B: uint8(getBlue(imgLoaded[y][x])),
				A: 255,
			})
		}
	}

	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
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

func main() {
	fmt.Println("Bienvenue sur notre application de filtres photo.")

	menu()

	file, err := os.Open(inputFile)

	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}

	defer file.Close()

	pixels, err := getImg(file)

	if err != nil {
		fmt.Println("Error: Image could not be decoded")
		os.Exit(1)
	}

	switch filterMenu {
	case 1:
		blackAndWhite(pixels)
	case 2:
		noiseReduction(pixels)
	}
	encode()

	fmt.Println("Filtre applique avec succes")
}

func noiseReduction(pixels [][]Pixel) {
	for y := 2; y < height-2; y++ {
		for x := 2; x < width-2; x++ {
			newRed := (getRed(pixels[y+1][x]) + getRed(pixels[y-1][x]) + getRed(pixels[y][x-1]) + getRed(pixels[y][x+1]) + getRed(pixels[y+1][x+1]) + getRed(pixels[y+1][x-1]) + getRed(pixels[y-1][x+1]) + getRed(pixels[y-1][x-1]) + 7*getRed(pixels[y][x])) / 15
			newGreen := (getGreen(pixels[y+1][x]) + getGreen(pixels[y-1][x]) + getGreen(pixels[y][x-1]) + getGreen(pixels[y][x+1]) + getGreen(pixels[y+1][x+1]) + getGreen(pixels[y+1][x-1]) + getGreen(pixels[y-1][x+1]) + getGreen(pixels[y-1][x-1]) + 7*getGreen(pixels[y][x])) / 15
			newBlue := (getBlue(pixels[y+1][x]) + getBlue(pixels[y-1][x]) + getBlue(pixels[y][x-1]) + getBlue(pixels[y][x+1]) + getBlue(pixels[y+1][x+1]) + getBlue(pixels[y+1][x-1]) + getBlue(pixels[y-1][x+1]) + getBlue(pixels[y-1][x-1]) + 7*getBlue(pixels[y][x])) / 15
			newAlpha := (getAlpha(pixels[y+1][x]) + getAlpha(pixels[y-1][x]) + getAlpha(pixels[y][x-1]) + getAlpha(pixels[y][x+1]) + getAlpha(pixels[y+1][x+1]) + getAlpha(pixels[y+1][x-1]) + getAlpha(pixels[y-1][x+1]) + getAlpha(pixels[y-1][x-1]) + 7*getAlpha(pixels[y][x])) / 15

			imgLoaded[y][x] = Pixel{newRed, newGreen, newBlue, newAlpha}
		}
	}
}

func blackAndWhite(img [][]Pixel) {
	for i := 0; i < height-1; i++ {
		for j := 0; j < width-1; j++ {
			pixel := img[i][j]

			newRed := (getRed(pixel) + getGreen(pixel) + getBlue(pixel)) / 3
			newGreen := newRed
			newBlue := newRed

			imgLoaded[i][j] = Pixel{newRed, newGreen, newBlue, getAlpha(pixel)}
		}
	}
}

func menu() {
	fmt.Println("\n")
	fmt.Println("Choisissez votre filtre :")
	fmt.Println("1 - Noir et Blanc")
	fmt.Println("2 - diminution du bruit")
	fmt.Println("0 - Annuler")
	fmt.Print("-> ")
	fmt.Scanf("%d", &filterMenu)

	if filterMenu == 0 {
		os.Exit(1)
	}

	fmt.Print("Qu'elle image voulez-vous traiter : ")
	fmt.Scanf("\r\n%s", &inputFile)
	fmt.Print("Donnez un nom à votre nouvelle image : ")
	fmt.Scanf("%s", &outputFile)
}
