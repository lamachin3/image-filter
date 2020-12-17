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

func encode(out chan Pixel) {
	finalImg := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			finalImg.Set(x, y, color.RGBA{
				R: uint8(getRed(<-out)),
				G: uint8(getGreen(<-out)),
				B: uint8(getBlue(<-out)),
				A: 255,
			})
		}
	}

	f, err := os.Create(outputFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := png.Encode(f, finalImg); err != nil {
		f.Close()
		log.Fatal(err)
	}

	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}

func main() {

	//imgEncoded := image.NewRGBA(image.Rect(0, 0, width, height))

	var inputChannel chan Pixel
	var feedbackChannel chan Pixel

	inputChannel = make(chan Pixel, 10)
	feedbackChannel = make(chan Pixel, 10)

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
		for nbRoutine := 0; nbRoutine < 10; nbRoutine++ {
			go blackAndWhite(inputChannel, feedbackChannel)
		}
	case 2:
		for nbRoutine := 0; nbRoutine < 10; nbRoutine++ {
			go noiseReduction(pixels)
		}
	}

	go feedInput(inputChannel, pixels)

	encode(feedbackChannel)

	fmt.Println("Filtre applique avec succes")
}

// OK ~
func feedInput(inp chan Pixel, pixels [][]Pixel) {
	for cptX := 1; cptX < width-1; cptX++ {
		for cptY := 1; cptY < height-1; cptY++ {
			toPush := pixels[cptX][cptY]
			inp <- toPush
		}
	}
	fmt.Printf("#DEBUG All Pushed\n")
}

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
}

func blackAndWhite(in chan Pixel, out chan Pixel) {
	pixel := <-in

	newRed := (getRed(pixel) + getGreen(pixel) + getBlue(pixel)) / 3
	newGreen := newRed
	newBlue := newRed

	out <- Pixel{newRed, newGreen, newBlue, getAlpha(pixel)}
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
