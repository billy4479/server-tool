package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

func WriteConfig() error {
	configPath, err := makeConfigFolder()
	if err != nil {
		return nil
	}

	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewEncoder(f).Encode(C)
}
