package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("this is file without os.Exit func call in main function, just in subfunction")
	exit()
}

func exit() {
	os.Exit(1)
}
