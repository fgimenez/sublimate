package runner

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const SublimateCfgFile = ".sublimate.yaml"

type Runner interface {
	Run() error
}

type App struct{}

type Proyect struct {
	summary *string
}

func (a *App) Run() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	file := filepath.Join(path, SublimateCfgFile)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return errors.New("missing sublimate config file " + SublimateCfgFile)
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	p := &Proyect{}
	err = yaml.Unmarshal(data, &p)
	if err != nil {
		return errors.New("non-valid sublimate config file " + SublimateCfgFile)
	}
	if p.summary == nil {
		return errors.New("sublimate config file missing summary " + SublimateCfgFile)
	}
	return nil
}
