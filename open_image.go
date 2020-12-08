package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
)

func main() {
	f, err := os.Open("./img.png")
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	imgINSA, imgName, err := image.Decode(f)
	fmt.Println(imgName)
	fmt.Println(err)
	
	//utiliser newRGBA avec un rectangle qui définit la taille de notre image
	blue := color.RGBA{0, 0, 255, 255}
	for i := 100; i < 105; i++ {
		for j := 100; j < 105; j++ {
			imgINSA.Set(i, j, blue)
		}
	}

	outFile, _ := os.Create("new_logo.png")
	defer outFile.Close()
	png.Encode(outFile, imgINSA)
}
