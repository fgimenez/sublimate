package main_test

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var sublimateBin string

func TestMain(m *testing.M) {
	// compile command/
	cmd := exec.Command("go", "get", "./cmd/sublimate/")
	cmd.Dir = filepath.Join(os.Getenv("GOPATH"), "src/github.com/fgimenez/sublimate")
	if err := cmd.Run(); err != nil {
		log.Fatalf("error building command %v", err)
	}
	defer os.Remove(sublimateBin)
	sublimateBin = filepath.Join(os.Getenv("GOPATH"), "bin/sublimate")
	os.Exit(m.Run())
}
