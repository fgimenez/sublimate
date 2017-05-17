package main_test

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	sublimateBin     = "/tmp/sublimate"
	sublimateCfgFile = ".sublimate.yaml"
)

var fixturesDir = filepath.Join(os.Getenv("GOPATH"), "src/github.com/fgimenez/sublimate/tests/fixtures")

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

func TestCfgFile(t *testing.T) {
	t.Run("cfg_not_found_shows_error", func(t *testing.T) {
		cmd := exec.Command(sublimateBin)
		cmd.Dir = filepath.Join(fixturesDir, "empty")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected error didn't happen")
		}
		if !strings.Contains(string(output), "missing sublimate config file "+sublimateCfgFile) {
			t.Fatalf("unexpected error received %s", output)
		}
	})
	for i := 1; i <= 2; i++ {
		t.Run("cfg_not_valid_shows_error", func(t *testing.T) {
			cmd := exec.Command(sublimateBin)
			cmd.Dir = filepath.Join(fixturesDir, fmt.Sprintf("non-valid-%d", i))
			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Fatalf("expected error didn't happen")
			}
			if !strings.Contains(string(output), "non-valid sublimate config file "+sublimateCfgFile) {
				t.Fatalf("unexpected error received %s", output)
			}
		})
	}
	t.Run("cfg_missing_summary_shows_error", func(t *testing.T) {
		cmd := exec.Command(sublimateBin)
		cmd.Dir = filepath.Join(fixturesDir, "missing-summary")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected error didn't happen")
		}
		if !strings.Contains(string(output), "sublimate config file missing summary "+sublimateCfgFile) {
			t.Fatalf("unexpected error received %s", output)
		}
	})
	t.Run("cfg_missing_contract_shows_error", func(t *testing.T) {
		cmd := exec.Command(sublimateBin)
		cmd.Dir = filepath.Join(fixturesDir, "missing-contract")
		output, err := cmd.CombinedOutput()
		if err == nil {
			t.Fatalf("expected error didn't happen")
		}
		if !strings.Contains(string(output), "sublimate config file missing contract "+sublimateCfgFile) {
			t.Fatalf("unexpected error received %s", output)
		}
	})

}
