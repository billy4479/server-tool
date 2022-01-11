package config

import (
	"os"

	"github.com/billy4479/server-tool/logger"
	"gopkg.in/yaml.v3"
)

func LoadConfig() error {
	configPath, err := makeConfigFolder()
	if err != nil {
		return nil
	}

	logger.L.Info.Printf("[+] Loading config at %s\n", configPath)

	f, err := os.Open(configPath)
	if err != nil {
		C = &defaultConfig
		return err
	}
	defer f.Close()

	C = NewConfig()
	err = yaml.NewDecoder(f).Decode(C)

	return err
}
