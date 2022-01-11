package server

import "github.com/billy4479/server-tool/manifest"

type ServerType uint8

const (
	Vanilla ServerType = iota
	Fabric
)

type Server struct {
	Name           string
	BaseDir        string
	Version        *manifest.VersionInfo
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
