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
		fmt.Printf("Error from reading: %v\n", err)
	}
	text := string(content)
	fmt.Println(text)

	// Write to host file from inside the container!
	f, err := os.OpenFile(hostFileToWriteLocation, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Printf("Error opening file to write: %v\n", err)
	}
	if _, err = f.WriteString("Hello from the container!\n"); err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
	}

	// Sleep
	time.Sleep(1 * time.Hour)
}
