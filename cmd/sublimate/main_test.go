package main_test

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const sublimateBin = "/tmp/sublimate"

func TestMain(m *testing.M) {
	// compile command/
	cmd := exec.Command("go", "build", "-o", sublimateBin, "./cmd/sublimate/")
	cmd.Dir = filepath.Join(os.Getenv("GOPATH"), "src/github.com/fgimenez/sublimate")
	if err := cmd.Run(); err != nil {
		log.Fatalf("error building command %v", err)
	}
	defer os.Remove(sublimateBin)
	os.Exit(m.Run())
}
