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

	for x := 0; x < width; x++ {
		var row []Pixel
		for y := 0; y < height; y++ {
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

	inputChannel = make(chan Pixel, 10)
	feedbackChannel = make(chan Pixel, 10)

	fmt.Println("Bienvenue sur notre application de filtres photo.")

	//menu()
	inputFile = "image1.png"
	outputFile = "output.png"
	filterMenu = 2

	file, err := os.Open(inputFile)

	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}

	defer file.Close()

	imgLoaded, err := getImg(file)
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
			go noiseReduction(imgLoaded, inputChannel, feedbackChannel, 1)
		}
	}

	go feedInput(inputChannel, imgLoaded)

	encode(feedbackChannel, img2Encode)

	fmt.Println("Filtre applique avec succes")

	createFile(img2Encode)

	fmt.Println("Fichier créé avec succes")
}

// OK ~
func feedInput(inp chan Pixel, pixels [][]Pixel) {
	for cptX := 0; cptX < width; cptX++ {
		for cptY := 0; cptY < height; cptY++ {
			toPush := pixels[cptX][cptY]
			inp <- toPush
		}
	}
	fmt.Printf("#DEBUG All Pushed\n")
}

func noiseReduction(img [][]Pixel, in chan Pixel, out chan Pixel, srdSize int) {
	for {
		var chgPixel Pixel
		cpt := 0
		pixel := <-in

		switch pixel.posX {
		case 0:
			switch pixel.posY {
			case 0:
				surroundMean(img, pixel, []int{0, srdSize, 0, srdSize}, &chgPixel, &cpt)
			case height - 1:
				surroundMean(img, pixel, []int{0, srdSize, srdSize, 0}, &chgPixel, &cpt)
			default:
				surroundMean(img, pixel, []int{0, srdSize, srdSize, srdSize}, &chgPixel, &cpt)
			}
		case width - 1:
			switch pixel.posY {
			case 0:
				surroundMean(img, pixel, []int{srdSize, 0, 0, srdSize}, &chgPixel, &cpt)
			case height - 1:
				surroundMean(img, pixel, []int{srdSize, 0, srdSize, 0}, &chgPixel, &cpt)
			default:
				surroundMean(img, pixel, []int{srdSize, 0, srdSize, srdSize}, &chgPixel, &cpt)
			}
		default:
			switch pixel.posY {
			case 0:
				surroundMean(img, pixel, []int{srdSize, srdSize, 0, srdSize}, &chgPixel, &cpt)
			case height - 1:
				surroundMean(img, pixel, []int{srdSize, srdSize, srdSize, 0}, &chgPixel, &cpt)
			default:
				surroundMean(img, pixel, []int{srdSize, srdSize, srdSize, srdSize}, &chgPixel, &cpt)
			}
		}

		out <- Pixel{chgPixel.R / cpt,
			chgPixel.G / cpt,
			chgPixel.B / cpt,
			chgPixel.A / cpt,
			pixel.posX, pixel.posY}
	}
}

func surroundMean(img [][]Pixel, pixel Pixel, srdSizes []int, chgPixel *Pixel, cpt *int) {
	for x := pixel.posX - srdSizes[0]; x <= pixel.posX+srdSizes[1]; x++ {
		for y := pixel.posY - srdSizes[2]; y <= pixel.posY+srdSizes[3]; y++ {
			ratio := 1
			if x == pixel.posX && y == pixel.posY {
				ratio = 7
			}
			chgPixel.R += img[x][y].R * ratio
			chgPixel.G += img[x][y].G * ratio
			chgPixel.B += img[x][y].B * ratio
			chgPixel.A += img[x][y].A * ratio
			*cpt += ratio
		}
	}
}

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
