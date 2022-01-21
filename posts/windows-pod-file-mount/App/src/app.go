package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	hostFileToReadLocation := os.Args[1]
	hostFileToWriteLocation := os.Args[2]

	fmt.Printf("host file to read: %s\n", hostFileToReadLocation)
	fmt.Printf("host file to write: %s\n", hostFileToWriteLocation)

	// Read host file from inside the container!
	content, err := ioutil.ReadFile(hostFileToReadLocation)
	if err != nil {
		panic(err)
	}
	text := string(content)
	fmt.Println(text)

	// Write to host file from inside the container!
	message := []byte("Hello from the container!")
	err = ioutil.WriteFile(hostFileToWriteLocation, message, 755)
	if err != nil {
		panic(err)
	}

	// Sleep
	time.Sleep(1 * time.Hour)
}
