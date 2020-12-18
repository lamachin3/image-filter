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
	// si il n'y a pas tous les arguments, redirection vers le manuel
	if len(os.Args) < 4 {
		help()
		os.Exit(0)
	}

	// on récupère les paramètres
	filter = os.Args[1]
	inputFile = os.Args[2]
	outputFile = os.Args[3]

	// channels pour les go routines
	var inputChannel chan Pixel
	var feedbackChannel chan Pixel

	inputChannel = make(chan Pixel, 10)
	feedbackChannel = make(chan Pixel, 10)

	fmt.Println("Bienvenue sur notre application de filtres photo.")

	// ouverture de l'image
	file, err := os.Open(inputFile)

	if err != nil {
		fmt.Println("Error: File could not be opened")
		os.Exit(1)
	}

	defer file.Close()

	// on crée une matrice de Pixel à partir de l'image
	imgLoaded, err := getImg(file)
	img2Encode := image.NewRGBA(image.Rect(0, 0, width, height))

	if err != nil {
		fmt.Println("Error: Image could not be decoded")
		os.Exit(1)
	}

	switch filter {
	// filtre noir et blanc
	case "1":
		for nbRoutine := 0; nbRoutine < 10; nbRoutine++ {
			// création de 10 go routines
			go blackAndWhite(inputChannel, feedbackChannel)
		}
	case "2":
	// filtre réducteur de bruit
		for nbRoutine := 0; nbRoutine < 10; nbRoutine++ {
			// création de 10 go routines
			go noiseReduction(imgLoaded, inputChannel, feedbackChannel, 1)
		}
	}

	// A DEFINIR !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	go feedInput(inputChannel, imgLoaded)

	// on récupère le résultat produit par une go routine
	// enregistrement de l'image traitée
	encode(feedbackChannel, img2Encode)

	fmt.Println("Filtre applique avec succes")

	// écriture de l'image traitée dans un fichier
	createFile(img2Encode)

	fmt.Println("Fichier créé avec succes")
}

// A DEFINIR !!!
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

		switch pixel.posX {
		// cas du bord gauche
		case 0:
			switch pixel.posY {
			// cas du bord haut
			case 0:
				surroundMean(img, pixel, []int{0, srdSize, 0, srdSize}, &chgPixel, &cpt)
			// cas du bord bas
			case height - 1:
				surroundMean(img, pixel, []int{0, srdSize, srdSize, 0}, &chgPixel, &cpt)
			// cas "normal"
			default:
				surroundMean(img, pixel, []int{0, srdSize, srdSize, srdSize}, &chgPixel, &cpt)
			}
		// cas du bord droit
		case width - 1:
			switch pixel.posY {
			// cas du bord haut
			case 0:
				surroundMean(img, pixel, []int{srdSize, 0, 0, srdSize}, &chgPixel, &cpt)
			// cas du bord bas
			case height - 1:
				surroundMean(img, pixel, []int{srdSize, 0, srdSize, 0}, &chgPixel, &cpt)
			// cas "normal"
			default:
				surroundMean(img, pixel, []int{srdSize, 0, srdSize, srdSize}, &chgPixel, &cpt)
			}
		// cas "normal"
		default:
			switch pixel.posY {
			// cas du bord haut
			case 0:
				surroundMean(img, pixel, []int{srdSize, srdSize, 0, srdSize}, &chgPixel, &cpt)
			// cas du bord bas
			case height - 1:
				surroundMean(img, pixel, []int{srdSize, srdSize, srdSize, 0}, &chgPixel, &cpt)
			// cas "normal"
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

// A DEFINIR !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
func surroundMean(img [][]Pixel, pixel Pixel, srdSizes []int, chgPixel *Pixel, cpt *int) {
	for x := pixel.posX - srdSizes[0]; x <= pixel.posX+srdSizes[1]; x++ {
		for y := pixel.posY - srdSizes[2]; y <= pixel.posY+srdSizes[3]; y++ {
			chgPixel.R += img[x][y].R
			chgPixel.G += img[x][y].G
			chgPixel.B += img[x][y].B
			chgPixel.A += img[x][y].A
			*cpt++
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
	fmt.Println("image-filter [filter-choice] [input-image] [output-image]\n")
	fmt.Println("filter-choice:\t1 - black and white\n\t\t2 - noise reduction")
}
