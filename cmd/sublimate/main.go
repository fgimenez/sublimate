package main

import (
	"log"

	"github.com/fgimenez/sublimate/pkg/runner"
)

func main() {
	a := &runner.App{}

	if err := a.Run(); err != nil {
		log.Fatalf("error : %v", err)
	}
}
