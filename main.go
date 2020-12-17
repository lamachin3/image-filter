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
	R    int
	G    int
	B    int
	A    int
	posX int
	posY int
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
			R, G, B, A := img.At(x, y).RGBA()
			row = append(row, rgbaToPixel(R, G, B, A, x, y))
		}
		imgLoaded = append(imgLoaded, row)
	}

	return imgLoaded, nil
}

//conversion uint32 -> uint8
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32, x int, y int) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257), x, y}
}

func encode(out chan Pixel, img2Encode *image.RGBA) {
	for i := 0; i <= (height-1)*(width-1); i++ {
		pixel := <-out
		//fmt.Print("(", pixel.posX, ";", pixel.posY, ";", uint8(pixel.A), ") /")
		img2Encode.Set(pixel.posX, pixel.posY, color.RGBA{
			R: uint8(pixel.R),
			G: uint8(pixel.G),
			B: uint8(pixel.B),
			A: uint8(pixel.A),
		})
	}
}

func createFile(img2Encode *image.RGBA) {
	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, img2Encode); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	var inputChannel chan Pixel
	var feedbackChannel chan Pixel

	inputChannel = make(chan Pixel, 1000)
	feedbackChannel = make(chan Pixel, 1000)

	fmt.Println("Bienvenue sur notre application de filtres photo.")

	//menu()
	inputFile = "image.png"
	outputFile = "output.png"
	filterMenu = 1

	file, err := os.Open(inputFile)

	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}

	defer file.Close()

	pixels, err := getImg(file)
	img2Encode := image.NewRGBA(image.Rect(0, 0, width, height))

	if err != nil {
		fmt.Println("Error: Image could not be decoded")
		os.Exit(1)
	}

	switch filterMenu {
	case 1:
		for nbRoutine := 0; nbRoutine < 10; nbRoutine++ {
			go blackAndWhite(inputChannel, feedbackChannel)
		}
	case 2:
		for nbRoutine := 0; nbRoutine < 10; nbRoutine++ {
			//go noiseReduction(pixels)
		}
	}

	go feedInput(inputChannel, pixels)

	encode(feedbackChannel, img2Encode)

	fmt.Println("Filtre applique avec succes")

	createFile(img2Encode)

	fmt.Println("Fichier créé avec succes")
}

// OK ~
func feedInput(inp chan Pixel, pixels [][]Pixel) {
	for cptX := 0; cptX < height; cptX++ {
		for cptY := 0; cptY < width; cptY++ {
			toPush := pixels[cptX][cptY]
			inp <- toPush
		}
	}
	fmt.Printf("#DEBUG All Pushed\n")
}

/*
func noiseReduction(pixels [][]Pixel) {
	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			newRed := (getRed(pixels[y+1][x]) + getRed(pixels[y-1][x]) + getRed(pixels[y][x-1]) + getRed(pixels[y][x+1]) + getRed(pixels[y+1][x+1]) + getRed(pixels[y+1][x-1]) + getRed(pixels[y-1][x+1]) + getRed(pixels[y-1][x-1]) + 7*getRed(pixels[y][x])) / 15
			newGreen := (getGreen(pixels[y+1][x]) + getGreen(pixels[y-1][x]) + getGreen(pixels[y][x-1]) + getGreen(pixels[y][x+1]) + getGreen(pixels[y+1][x+1]) + getGreen(pixels[y+1][x-1]) + getGreen(pixels[y-1][x+1]) + getGreen(pixels[y-1][x-1]) + 7*getGreen(pixels[y][x])) / 15
			newBlue := (getBlue(pixels[y+1][x]) + getBlue(pixels[y-1][x]) + getBlue(pixels[y][x-1]) + getBlue(pixels[y][x+1]) + getBlue(pixels[y+1][x+1]) + getBlue(pixels[y+1][x-1]) + getBlue(pixels[y-1][x+1]) + getBlue(pixels[y-1][x-1]) + 7*getBlue(pixels[y][x])) / 15
			newAlpha := (getAlpha(pixels[y+1][x]) + getAlpha(pixels[y-1][x]) + getAlpha(pixels[y][x-1]) + getAlpha(pixels[y][x+1]) + getAlpha(pixels[y+1][x+1]) + getAlpha(pixels[y+1][x-1]) + getAlpha(pixels[y-1][x+1]) + getAlpha(pixels[y-1][x-1]) + 7*getAlpha(pixels[y][x])) / 15

			imgLoaded[y][x] = Pixel{newRed, newGreen, newBlue, newAlpha}
		}
	}
}*/

func blackAndWhite(in chan Pixel, out chan Pixel) {
	for {
		pixel := <-in

		newRed := (pixel.R + pixel.G + pixel.B) / 3
		newGreen := newRed
		newBlue := newRed

		out <- Pixel{newRed, newGreen, newBlue, pixel.A, pixel.posX, pixel.posY}
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
	fmt.Scanf("\r\n%s", &outputFile)
}
