package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("this is file with os.Exit func call in main function")
	os.Exit(1) // want `os.Exit called from main func`
}
