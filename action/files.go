package action

import (
	"errors"
	"os"
	"path"

	"gopkg.in/yaml.v3"
)

type RequestInfo struct {
	Version     string   `yaml:"version"`
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Method      string   `yaml:"method"`
	Path        string   `yaml:"path"`
	Headers     []string `yaml:"headers"`
	Body        string   `yaml:"body"`
}

func WriteRequestFiles(cfgPath, app string, req *RequestInfo) error {
	data, err := yaml.Marshal(req)
	if err != nil {
		return errors.New("failed to marshal data")
	}
	path := path.Join(AppPath(cfgPath, app), req.Name+".yml")
	return os.WriteFile(path, data, 0700)
}

type AppInfo struct {
	Version     string `yaml:"version"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Host        string `yaml:"host"`
}

func WriteAppFiles(cfgPath string, app *AppInfo) error {
	data, err := yaml.Marshal(app)
	if err != nil {
		return errors.New("failed to marshal data")
	}
	err = os.MkdirAll(AppPath(cfgPath, app.Name), 0700)
	if err != nil {
		return err
	}
	err = os.WriteFile(AppInfoFilePath(cfgPath, app.Name), data, 0700)
	if err != nil {
		os.Remove(AppPath(cfgPath, app.Name))
		return err
	}
	return nil
}
