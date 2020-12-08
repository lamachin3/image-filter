package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
)

func chargementImage(nomFichier string) image.Image {
	f, err := os.Open(nomFichier)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	img, err := jpeg.Decode(f)
	if err != nil {
		log.Fatal(err)
	}
	return img
}

func affichPixel(img image.Image, x int, y int) {
	//fmt.Println(img.At(10, 10))
	rgba := img.At(x, y)
	//fmt.Printf("[X : %d Y : %v] R : %v, G : %v, B : %v\n", x, y, r, g, b)
	fmt.Printf("rgba : %d", rgba)
}

func main() {
	cheminIMG := "./img.jpg"
	fmt.Println("Chargement image...")
	img := chargementImage(cheminIMG)
	fmt.Println("Affichage pixel...")
	affichPixel(img, 100, 100)
}
