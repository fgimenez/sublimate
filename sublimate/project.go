package sublimate

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

type Project struct {
	Summary  string
	Contract string
	Script   string
}

func newProject(path string) (*Project, error) {
	file := filepath.Join(path, sublimateCfgFile)
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return nil, errors.New("missing sublimate config file " + sublimateCfgFile)
	}
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	p := &Project{}
	err = yaml.Unmarshal(data, p)
	if err != nil {
		return nil, errors.New("non-valid sublimate config file " + sublimateCfgFile)
	}
	if err := p.validate(path); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Project) validate(path string) error {
	if p.Summary == "" {
		return errors.New("sublimate config file missing summary " + sublimateCfgFile)
	}
	if p.Contract == "" {
		return errors.New("sublimate config file missing contract " + sublimateCfgFile)
	}
	fpath := filepath.Join(path, p.Contract)
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return errors.New("contract file not found " + p.Contract)
	}
	p.Contract = fpath
	if p.Script == "" {
		return errors.New("sublimate config file missing script " + sublimateCfgFile)
	}
	return nil
}
