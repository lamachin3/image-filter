package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	content, err := ioutil.ReadFile("lorem.txt")

	if err != nil {
		log.Fatal(err)
	}
	para := strings.Split(string(content), "\n")

	for i := 0; i < len(para)-1; i++ {
		words := strings.Split(para[i], " ")
		fmt.Println(words[len(words)-1])
	}
}
