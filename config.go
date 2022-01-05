package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
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
	config        *Config = nil
)

func NewConfig() *Config {
	c := new(Config)
	{
		c.Application.Quiet = false
		c.Application.WorkingDir = "."
		c.Application.CacheDir = ""
	}
	{
		c.Minecraft.Quiet = false
		c.Minecraft.NoGUI = false
		c.Minecraft.NoEULA = false
	}
	{
		c.Java.ExecutableOverride = ""
		c.Java.Memory.Amount = 6
		c.Java.Memory.Gigabytes = true
		c.Java.Flags.ExtraFlags = nil
		c.Java.Flags.OverrideDefault = false
	}
	{
		c.Git.Disable = false
		c.Git.DisableGithubIntegration = false
		c.Git.Overrides.Enable = false
		c.Git.Overrides.CustomPreCommands = nil
		c.Git.Overrides.CustomPostCommands = nil
	}
	return c
}

func makeConfigFolder() (configPath string, err error) {
	configPathOverride := os.Getenv("CONFIG_PATH")
	configDir := ""
	if configPathOverride != "" {
		configPath = configPathOverride
		configDir = filepath.Dir(configPath)
	} else {
		configDir, err = os.UserConfigDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(configDir, progName)
		configPath = filepath.Join(configDir, progName+".yml")
	}

	return configPath, os.MkdirAll(configDir, 0700)
}

func writeConfig() error {
	configPath, err := makeConfigFolder()
	if err != nil {
		return nil
	}

	f, err := os.Create(configPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return yaml.NewEncoder(f).Encode(config)
}

func loadConfig() error {
	configPath, err := makeConfigFolder()
	if err != nil {
		return nil
	}

	Info.Printf("[+] Loading config at %s\n", configPath)

	f, err := os.Open(configPath)
	if err != nil {
		config = &defaultConfig
		return err
	}
	defer f.Close()

	config = NewConfig()
	err = yaml.NewDecoder(f).Decode(config)

	return err
}
