package main

//importation des modules
import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"os"
	"strconv"
)

//définition d'un type Pixel
// RGBA : Rouge Vert BLeu Transparence
// coordonnées posX posY du pixel
type Pixel struct {
	R    int
	G    int
	B    int
	A    int
	posX int
	posY int
}

// paramètres d'éxécution du programme
var filter, inputFile, outputFile string

// dimension de l'image
var height, width = 0, 0

// matrice de Pixel contenant l'image à traiter
var imgLoaded [][]Pixel

// crée une matrice à partir d'une image
// récupère les valeurs RGBA de chaque pixel
func getImg(file io.Reader) ([][]Pixel, error) {
	// ouverture du fichier
	img, _, err := image.Decode(file)

	if err != nil {
		return nil, err
	}

	// définition de la hauteur de la largueur de l'image à traiter
	bounds := img.Bounds()
	width, height = bounds.Max.X, bounds.Max.Y

	// parcours en 2 dimensions
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

//conversion uint32 -> uint8 pour avoir des valeurs comprises entre 0 et 255
// 0,0,0 (noir) -> 255,255,255 (blanc)
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32, x int, y int) Pixel {
	return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257), x, y}
}

// place la valeur du pixel traité (par un channel) aux bonnes coordonnées dans une nouvelle image
func encode(out chan Pixel, img2Encode *image.RGBA) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pixel := <-out
			img2Encode.Set(pixel.posX, pixel.posY, color.RGBA{
				R: uint8(pixel.R),
				G: uint8(pixel.G),
				B: uint8(pixel.B),
				A: uint8(pixel.A),
			})
		}
	}
}

// écriture d'une image dans un nouveau fichier
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

// boucle principale
func main() {
	if len(os.Args) < 4 {
		help()
		os.Exit(0)
	}

	filter = os.Args[1]
	inputFile = os.Args[2]
	outputFile = os.Args[3]

	var inputChannel chan Pixel
	var feedbackChannel chan Pixel

	inputChannel = make(chan Pixel, 10)
	feedbackChannel = make(chan Pixel, 10)

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

	switch filter {
	case "1":
		for nbRoutine := 0; nbRoutine < 10; nbRoutine++ {
			go blackAndWhite(inputChannel, feedbackChannel)
		}
	case "2":
		srdSize := 1
		if len(os.Args) == 5 {
			srdSize, err = strconv.Atoi(os.Args[4])
		}
		for nbRoutine := 0; nbRoutine < 10; nbRoutine++ {
			go noiseReduction(imgLoaded, inputChannel, feedbackChannel, srdSize)
		}
	}

	go feedInput(inputChannel, imgLoaded)

	encode(feedbackChannel, img2Encode)

	fmt.Println("Filtre applique avec succes")

	createFile(img2Encode)

	fmt.Println("Fichier créé avec succes")
}

// fonction qui nous permet d'ajouter chaque pixel dans le channel d'entrée des go routines
func feedInput(inp chan Pixel, pixels [][]Pixel) {
	for cptX := 0; cptX < width; cptX++ {
		for cptY := 0; cptY < height; cptY++ {
			toPush := pixels[cptX][cptY]
			inp <- toPush
		}
	}
}

// fonction réducteur de bruit
func noiseReduction(img [][]Pixel, in chan Pixel, out chan Pixel, srdSize int) {
	for {
		var chgPixel Pixel
		cpt := 0
		pixel := <-in
		posX, posY := pixel.posX, pixel.posY

		switch {
		case posX-srdSize < 0:
			switch {
			case posY-srdSize < 0:
				surroundMean(img, pixel, []int{posX, srdSize, posY, srdSize}, &chgPixel, &cpt)
			case posY+srdSize > height-1:
				surroundMean(img, pixel, []int{posX, srdSize, srdSize, height - posY - 1}, &chgPixel, &cpt)
			default:
				surroundMean(img, pixel, []int{posX, srdSize, srdSize, srdSize}, &chgPixel, &cpt)
			}
		case posX+srdSize > width-1:
			switch {
			case posY-srdSize < 0:
				surroundMean(img, pixel, []int{srdSize, width - posX - 1, posY, srdSize}, &chgPixel, &cpt)
			case posY+srdSize > height-1:
				surroundMean(img, pixel, []int{srdSize, width - posX - 1, srdSize, height - posY - 1}, &chgPixel, &cpt)
			default:
				surroundMean(img, pixel, []int{srdSize, width - posX - 1, srdSize, srdSize}, &chgPixel, &cpt)
			}
		default:
			switch {
			case posY-srdSize < 0:
				surroundMean(img, pixel, []int{srdSize, srdSize, posY, srdSize}, &chgPixel, &cpt)
			case posY+srdSize > height-1:
				surroundMean(img, pixel, []int{srdSize, srdSize, srdSize, height - posY - 1}, &chgPixel, &cpt)
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

// Fonction qui pour chaque pixel de l'image, fait la moyenne des 4 composantes entre les pixels autours du pixel courant
// srdSizes permet de définir la taille du carré qui va être pris en compte dans le calcul de moyenne
func surroundMean(img [][]Pixel, pixel Pixel, srdSizes []int, chgPixel *Pixel, cpt *int) {
	for x := pixel.posX - srdSizes[0]; x <= pixel.posX+srdSizes[1]; x++ {
		for y := pixel.posY - srdSizes[2]; y <= pixel.posY+srdSizes[3]; y++ {
			ratio := 1
			if x == pixel.posX && y == pixel.posY {
				ratio = 20
			}
			chgPixel.R += img[x][y].R * ratio
			chgPixel.G += img[x][y].G * ratio
			chgPixel.B += img[x][y].B * ratio
			chgPixel.A += img[x][y].A * ratio
			*cpt += ratio
		}
	}
}

// filtre noir et blanc
func blackAndWhite(in chan Pixel, out chan Pixel) {
	for {
		pixel := <-in

		newRed := (pixel.R + pixel.G + pixel.B) / 3
		newGreen := newRed
		newBlue := newRed

		out <- Pixel{newRed, newGreen, newBlue, pixel.A, pixel.posX, pixel.posY}
	}
}

// appelé si le programme est exécuté sans les 3 paramètres demandés
func help() {
	fmt.Println("\nMANUAL\n")
	fmt.Println("Pour appliqué un filtre, lancer le programme avec l'instruction suivante\n")
	fmt.Println("main.go [filter-choice] [input-image] [output-image] [OPTION] {noise-reduction-level}\n")
	fmt.Println("filter-choice:\t1 - black and white\n\t\t2 - noise reduction")
}
