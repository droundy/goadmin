package main

import (
	"fmt"
	"github.com/droundy/goopt"
)

func main() {
	goopt.Parse(func() []string { return []string{} })

	fmt.Println("Hello world!")
}
