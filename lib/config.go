package lib

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Application struct {
		AutoUpdate bool
		WorkingDir string
		CacheDir   string
	}
	Minecraft struct {
		Quiet  bool
		GUI    bool
		NoEULA bool
		Memory uint
	}
	Git struct {
		Enable      bool
		UseLockFile bool
	}
	UseSystemJava bool
}

var (
	defaultConfig         = *NewConfig()
	C             *Config = nil
)

func GetConfigPath() (configPath string, configDir string, err error) {
	configPathOverride := os.Getenv("CONFIG_PATH")
	if configPathOverride != "" {
		configPath = configPathOverride
		configDir = filepath.Dir(configPath)
	} else {
		configDir, err = os.UserConfigDir()
		if err != nil {
			return "", "", err
		}
		configDir = filepath.Join(configDir, ProgName)
		configPath = filepath.Join(configDir, ProgName+".yml")
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

func NewConfig() *Config {
	c := new(Config)
	{
		c.Application.WorkingDir = "."
		c.Application.CacheDir = ""
		c.Application.AutoUpdate = true
	}
	{
		c.Minecraft.Quiet = false
		c.Minecraft.GUI = false
		c.Minecraft.NoEULA = false
		c.Minecraft.Memory = 6 * 1024
	}
	{
		c.Git.Enable = true
		c.Git.UseLockFile = true
	}
	c.UseSystemJava = false
	return c
}

func LoadConfig() error {
	configPath, err := makeConfigFolder()
	if err != nil {
		return nil
	}

	// The logger is not yet initialized at this point
	// L.Debug.Printf("Loading config at %s\n", configPath)

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

	encoder := yaml.NewEncoder(f)
	encoder.SetIndent(2)
	return encoder.Encode(C)
}
