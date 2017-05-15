package main

import "log"

type runner interface {
	Run() error
}

type app struct{}

func (a *app) Run() error {
	return nil
}

func main() {
	a := &app{}

	if err := a.Run(); err != nil {
		log.Fatalf("error run app: %v", err)
	}
}
