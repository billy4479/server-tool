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

const adoptiumApiUrl = "https://api.adoptium.net/v3/assets/latest/%d/hotspot?release=latest&jvm_impl=hotspot&vendor=adoptium"

func javaDir() string     { return path.Join(config.C.Application.CacheDir, "java") }
func javaExePath() string { return path.Join("bin", javaExeName()) }
