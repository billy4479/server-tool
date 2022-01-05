package main

import (
	"errors"
	"os"
	"path/filepath"
)

type ServerType uint8

const (
	Vanilla ServerType = iota
	Fabric
)

type Server struct {
	Name           string
	BaseDir        string
	Version        *VersionInfo
	Type           ServerType
	HasGit         bool
	HasStartScript bool
	Start          func() error
}

const (
	FabricJarName    = "fabric-server-launch.jar"
	VanillaJarName   = "server.jar"
	GitDirectoryName = ".git"
)

func findServers() ([]Server, error) {
	serverDirs, err := os.ReadDir(config.Application.WorkingDir)
	if err != nil {
		return nil, err
	}

	servers := []Server{}

	for _, e := range serverDirs {
		var s Server
		s.BaseDir = filepath.Join(config.Application.WorkingDir, e.Name())

		if !e.IsDir() {
			continue
		}

		entries, err := os.ReadDir(s.BaseDir)
		if err != nil {
			return nil, err
		}

		s.Type = Vanilla
		for _, entry := range entries {
			if !entry.IsDir() {
				if entry.Name() == VanillaJarName {
					possibleServerJar := filepath.Join(s.BaseDir, entry.Name())
					err = detectServerVersion(possibleServerJar, &s)
					if err != nil {
						return nil, err
					}

					if s.Version == nil {
						// server.jar is not a Minecraft server
						break
					}
				} else if entry.Name() == FabricJarName {
					s.Type = Fabric
				} else if entry.Name() == config.StartScript.Name {
					s.HasStartScript = !config.StartScript.Disable
				}
			} else {
				if entry.Name() == GitDirectoryName {
					s.HasGit = !config.Git.Disable
				}
			}
		}
		if s.Version == nil && !s.HasStartScript {
			continue
		}

		s.Name = e.Name()
		setStartFn(&s)

		servers = append(servers, s)
	}

	if len(servers) == 0 {
		return servers, errors.New("No server were found!")
	}

	return servers, nil
}
