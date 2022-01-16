package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	configMapFileLocation := os.Args[1]

	fmt.Printf("configMapFileLocation: %s\n", configMapFileLocation)

	// Read mounted config map now as a file!
	fmt.Println("Reading the config map as a file!")
	content, err := ioutil.ReadFile(configMapFileLocation)
	if err != nil {
		panic(err)
	}
	text := string(content)
	fmt.Println(text)

	// Sleep
	time.Sleep(1 * time.Hour)
}
