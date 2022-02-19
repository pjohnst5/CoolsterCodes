package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Hello from the HostProcess pod!")

	// Sleep
	time.Sleep(1 * time.Hour)
}
