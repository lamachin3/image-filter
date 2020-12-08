package main

import (
    "fmt"
    "image"
    "image/png"
    "os"
    "io"
    "image/color"
    "log"
)

//type Pixel
type Pixel struct {
    R int
    G int
    B int
    A int
}

var height, width = 0, 0
var pixels [][]Pixel

// crée une matrice à partir d'une image
// récupère les valeurs RGBA de chaque pixel
func getPixels(file io.Reader) ([][]Pixel, error) {
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
        pixels = append(pixels, row)
    }
    return pixels, nil
}

//conversion uint32 -> uint8
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) Pixel {
    return Pixel{int(r / 257), int(g / 257), int(b / 257), int(a / 257)}
}


func main() {
    image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)

    file, err := os.Open("Tapping-Noise-in-Attic-at-Night.png")

    if err != nil {
        fmt.Println("Error: File could not be opened")
        os.Exit(1)
    }

    defer file.Close()

    pixels, err := getPixels(file)

    if err != nil {
        fmt.Println("Error: Image could not be decoded")
        os.Exit(1)
    }

    pixels[100][100] = Pixel{255, 0, 0, 255}
    fmt.Println(pixels[100][100])
    fmt.Println(getRed(pixels[100][100]))

    encode()

}


func encode() {
    // Create a colored image of the given width and height.
    img := image.NewRGBA(image.Rect(0, 0, width, height))

    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            img.Set(x, y, color.RGBA{
                R: uint8(getRed(pixels[y][x])),
                G: uint8(getGreen(pixels[y][x])),
                B: uint8(getBlue(pixels[y][x])),
                A: 255,
            })
        }
    }

    f, err := os.Create("image.png")
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



func getRed(lePixel Pixel) int{
    return lePixel.R
}

func getGreen(lePixel Pixel) int{
    return lePixel.G
}

func getBlue(lePixel Pixel) int{
    return lePixel.B
}


