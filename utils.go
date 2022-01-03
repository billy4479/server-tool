package main

import (
	"fmt"
	"os"
	"path/filepath"
)

var (
	cacheDir  string
	configDir string
)

func populateDataDirs() (err error) {
	cacheDir, err = os.UserCacheDir()
	if err != nil {
		return err
	}
	cacheDir = filepath.Join(cacheDir, "server-tool")
	if err = os.MkdirAll(cacheDir, 0700); err != nil {
		return err
	}

	configDir, err = os.UserConfigDir()
	if err != nil {
		return err
	}
	configDir = filepath.Join(configDir, "server-tool")
	if err = os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	return nil
}

func getWorkDir() string {
	baseDir := os.Getenv("BASE_DIR")
	if baseDir != "" {
		return baseDir
	}

	return "."
}

func makeServersMenuItem(servers []Server) []Option {
	result := []Option{}

	for _, s := range servers {
		desc := fmt.Sprintf("\"%s\" (", s.Name)
		if s.Version == nil {
			desc += "?? on ??"
		} else {
			desc += fmt.Sprintf("%s on ", s.Version.ID)
			switch s.Type {
			case Vanilla:
				desc += "Vanilla"
			case Fabric:
				desc += "Fabric"
			case Forge:
				desc += "Forge"
			case Paper:
				desc += "PaperMC"
			}
		}

		if s.HasGit {
			desc += " - Git"
		}
		if s.HasStartScript {
			desc += ", Start Script"
		}

		desc += ")"

		result = append(result, Option{
			Description: desc,
			Action:      s.Start,
		})
	}

	return result
}
