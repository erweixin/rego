package main

import (
	"fmt"

	rego "github.com/erweixin/rego"
)

func main() {
	if err := rego.Run(GalleryApp); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
