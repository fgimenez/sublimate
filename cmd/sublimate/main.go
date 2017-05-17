package main

import (
	"log"

	"github.com/fgimenez/sublimate/sublimate"
)

func main() {
	a := &sublimate.App{}

	if err := a.Run(); err != nil {
		log.Fatalf("error : %v", err)
	}
}
