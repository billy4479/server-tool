package main

import (
	"os"
	"path/filepath"
)

type ServerType uint8

const (
	Vanilla ServerType = iota
	Fabric
	Forge
	Paper
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
	StartScriptName  = "start.sh"
	FabricJarName    = "fabric-server-launch.jar"
	VanillaJarName   = "server.jar"
	GitDirectoryName = ".git"
)

func findServers() ([]Server, error) {
	serverDirs, err := os.ReadDir(getWorkDir())
	if err != nil {
		return nil, err
	}

	servers := []Server{}

	for _, e := range serverDirs {
		if !e.IsDir() ||
			e.Name() == ".server-tools" {
			continue
		}

		var s Server
		s.BaseDir = filepath.Join(getWorkDir(), e.Name())

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
				} else if entry.Name() == StartScriptName {
					s.HasStartScript = true
				}
			} else {
				if entry.Name() == GitDirectoryName {
					s.HasGit = true
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

	return servers, nil
}
