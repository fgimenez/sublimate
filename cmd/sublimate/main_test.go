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

func commonTest(t *testing.T, path, errorMessage string) {
	cmd := exec.Command(sublimateBin)
	cmd.Dir = filepath.Join(fixturesDir, path)
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected error didn't happen")
	}
	if !strings.Contains(string(output), errorMessage) {
		t.Fatalf("unexpected error received %s", output)
	}

}

func TestCfgFile(t *testing.T) {
	t.Run("cfg_not_found_shows_error", func(t *testing.T) {
		commonTest(t, "empty", "missing sublimate config file "+sublimateCfgFile)
	})
	for i := 1; i <= 2; i++ {
		t.Run("cfg_not_valid_shows_error", func(t *testing.T) {
			commonTest(t, fmt.Sprintf("non-valid-%d", i), "non-valid sublimate config file "+sublimateCfgFile)
		})
	}
	t.Run("cfg_missing_summary_shows_error", func(t *testing.T) {
		commonTest(t, "missing-summary", "sublimate config file missing summary "+sublimateCfgFile)
	})
	t.Run("cfg_missing_contract_shows_error", func(t *testing.T) {
		commonTest(t, "missing-contract", "sublimate config file missing contract "+sublimateCfgFile)
	})
	t.Run("cfg_non_file_contract_shows_error", func(t *testing.T) {
		commonTest(t, "non-file-contract", "contract file not found")
	})
	t.Run("cfg_missing_script_shows_error", func(t *testing.T) {
		commonTest(t, "missing-script", "sublimate config file missing script "+sublimateCfgFile)
	})
}
