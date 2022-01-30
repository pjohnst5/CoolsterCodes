package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	configMapFilePath := os.Args[1]

	fmt.Printf("configMap file path: %s\n", configMapFilePath)

	// Read the configMap as a file from inside the container!
	content, err := ioutil.ReadFile(configMapFilePath)
	if err != nil {
		fmt.Printf("Error from reading: %v\n", err)
	}
	text := string(content)
	fmt.Println(text)

	// Sleep
	time.Sleep(1 * time.Hour)
}
