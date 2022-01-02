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
	Version        *VersionInfo
	Type           ServerType
	HasGit         bool
	HasStartScript bool
	Start          func() error
}

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

		serverDir := filepath.Join(getWorkDir(), e.Name())

		entries, err := os.ReadDir(serverDir)
		if err != nil {
			return nil, err
		}

		var s Server
		s.Type = Vanilla
		for _, entry := range entries {
			if !entry.IsDir() {
				if entry.Name() == "server.jar" {
					possibleServerJar := filepath.Join(serverDir, entry.Name())
					err = detectServerVersion(possibleServerJar, &s)
					if err != nil {
						return nil, err
					}

					if s.Version == nil {
						// server.jar is not a Minecraft server
						break
					}
				} else if entry.Name() == "fabric-server-launch.jar" {
					s.Type = Fabric
				} else if entry.Name() == "start.sh" {
					s.HasStartScript = true
				}
			} else {
				if entry.Name() == ".git" {
					s.HasGit = true
				}
			}
		}
		if s.Version == nil && !s.HasStartScript {
			continue
		}

		s.Name = e.Name()
		s.Start = func() error {
			if s.Version != nil {
				Info.Printf("[+] %s requires Java %d\n", s.Name, s.Version.JavaVersion)
				javaExe, err := ensureJavaIsInstalled(s.Version.JavaVersion)
				if err != nil {
					return err
				}
				Ok.Printf("[+] Java executable: %s\n", javaExe)
			} else {
				Info.Printf("[+] %s has a startup script!\n", s.Name)
			}

			return nil
		}

		servers = append(servers, s)
	}

	return servers, nil
}
