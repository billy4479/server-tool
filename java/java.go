package java

import (
	"path"
	"runtime"

	"github.com/billy4479/server-tool/config"
)

func javaExeName() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

const adoptiumApiUrl = "https://api.adoptium.net/v3/assets/latest/%d/hotspot?os=%s&architecture=x64&image_type=jre"

func javaDir() string     { return path.Join(config.C.Application.CacheDir, "java") }
func javaExePath() string { return path.Join("bin", javaExeName()) }
