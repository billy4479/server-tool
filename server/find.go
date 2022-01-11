package server

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/billy4479/server-tool/config"
)

func FindServers() ([]Server, error) {
	serverDirs, err := os.ReadDir(config.C.Application.WorkingDir)
	if err != nil {
		return nil, err
	}

	servers := []Server{}

	for _, e := range serverDirs {
		var s Server
		s.BaseDir = filepath.Join(config.C.Application.WorkingDir, e.Name())

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
				} else if entry.Name() == config.C.StartScript.Name {
					s.HasStartScript = !config.C.StartScript.Disable
				}
			} else {
				if entry.Name() == GitDirectoryName {
					s.HasGit = !config.C.Git.Disable
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
