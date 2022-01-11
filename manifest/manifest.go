package manifest

import (
	"path/filepath"
	"time"

	"github.com/billy4479/server-tool/config"
)

type VersionManifestJSON struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		URL         string `json:"url"`
		Time        string `json:"time"`
		ReleaseTime string `json:"releaseTime"`
	} `json:"versions"`
}

type VersionInfo struct {
	ID          string
	JarURL      string
	JavaVersion int
	SHA         string
}

const (
	versionManifestURL = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
	expireTime         = 24 * time.Hour
)

func manifestDir() string     { return filepath.Join(config.C.Application.CacheDir, "manifest") }
func versionManifest() string { return filepath.Join(manifestDir(), "version_manifest.json") }
func versionInfos() string    { return filepath.Join(manifestDir(), "version_infos.json") }
