package config

import (
	"os"
	"path/filepath"

	"github.com/billy4479/server-tool/utils"
)

type Config struct {
	Application struct {
		Quiet      bool
		WorkingDir string
		CacheDir   string
	}
	Minecraft struct {
		Quiet  bool
		NoGUI  bool
		NoEULA bool
	}
	Java struct {
		ExecutableOverride string
		Memory             struct {
			Amount    uint
			Gigabytes bool
		}
		Flags struct {
			ExtraFlags      []string
			OverrideDefault bool
		}
	}
	Git struct {
		Disable                  bool
		DisableGithubIntegration bool
		UseLockFile              bool
		Overrides                struct {
			Enable             bool
			CustomPreCommands  [][]string
			CustomPostCommands [][]string
		}
	}
	StartScript struct {
		Disable bool
		Name    string
	}
}

var (
	defaultConfig         = *NewConfig()
	C             *Config = nil
)

func GetConfigPath() (configPath string, configDir string, err error) {
	configPathOverride := os.Getenv("CONFIG_PATH")
	configDir = ""
	if configPathOverride != "" {
		configPath = configPathOverride
		configDir = filepath.Dir(configPath)
	} else {
		configDir, err = os.UserConfigDir()
		if err != nil {
			return "", "", err
		}
		configDir = filepath.Join(configDir, utils.ProgName)
		configPath = filepath.Join(configDir, utils.ProgName+".yml")
	}

	return
}

func makeConfigFolder() (configPath string, err error) {
	configPath, configDir, err := GetConfigPath()
	if err != nil {
		return
	}

	return configPath, os.MkdirAll(configDir, 0755)
}
