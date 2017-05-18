package sublimate

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const sublimateCfgFile = ".sublimate.yaml"

type App struct{}

type Proyect struct {
	Summary  string
	Contract string
	Script   string
}

func (a *App) Run() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	file := filepath.Join(path, sublimateCfgFile)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return errors.New("missing sublimate config file " + sublimateCfgFile)
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	p := &Proyect{}
	err = yaml.Unmarshal(data, p)
	if err != nil {
		return errors.New("non-valid sublimate config file " + sublimateCfgFile)
	}
	if p.Summary == "" {
		return errors.New("sublimate config file missing summary " + sublimateCfgFile)
	}
	if p.Contract == "" {
		return errors.New("sublimate config file missing contract " + sublimateCfgFile)
	}
	if _, err := os.Stat(filepath.Join(path, p.Contract)); os.IsNotExist(err) {
		return errors.New("contract file not found " + p.Contract)
	}
	if p.Script == "" {
		return errors.New("sublimate config file missing script " + sublimateCfgFile)
	}
	return nil
}
