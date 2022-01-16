package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

func main() {
	fmt.Println("hello world")
	hostFile := os.Args[1]

	fmt.Printf("hostFile: %s\n", hostFile)

	// Read host file contents
	content, err := ioutil.ReadFile(hostFile)
	if err != nil {
		panic(err)
	}
	text := string(content)
	fmt.Println(text)

	// Append to host file contents
	f, err := os.OpenFile(hostFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	if _, err = f.WriteString("\nHello from the pod"); err != nil {
		panic(err)
	}
	f.Close()

	// Read host file contents again
	content2, err := ioutil.ReadFile(hostFile)
	if err != nil {
		panic(err)
	}
	text2 := string(content2)
	fmt.Println(text2)

	// Sleep
	time.Sleep(1 * time.Hour)
}
