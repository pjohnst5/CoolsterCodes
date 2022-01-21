package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	hostFileLocation := os.Args[1]

	fmt.Printf("host file: %s\n", hostFileLocation)

	// Read host file from inside the container!
	content, err := ioutil.ReadFile(hostFileLocation)
	if err != nil {
		panic(err)
	}
	text := string(content)
	fmt.Println(text)

	// Write to host file from inside the container!
	f, err := os.OpenFile(hostFileLocation, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	if _, err = f.WriteString("Hello from the container!\n"); err != nil {
		panic(err)
	}

	f.Close()

	// Sleep
	time.Sleep(1 * time.Hour)
}
