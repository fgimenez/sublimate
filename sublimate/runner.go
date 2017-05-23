package sublimate

import (
	"os"
)

const sublimateCfgFile = ".sublimate.yaml"

type App struct{}

func (a *App) Run() error {
	path, err := os.Getwd()
	if err != nil {
		return err
	}
	p, err := newProject(path)
	if err != nil {
		return err
	}

	g := &geth{}
	if err := g.run(p); err != nil {
		return err
	}

	return nil
}
