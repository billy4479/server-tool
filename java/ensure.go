package java

import (
	"fmt"
	"os"
	"path"

	"github.com/billy4479/server-tool/config"
	"github.com/billy4479/server-tool/logger"
)

func EnsureJavaIsInstalled(javaVersion int) (string, error) {
	javaVersionString := fmt.Sprint(javaVersion)
	err := os.MkdirAll(javaDir(), 0700)
	if err != nil {
		return "", nil
	}

	fullExePath := path.Join(javaDir(), javaVersionString, javaExePath())

	entries, err := os.ReadDir(javaDir())
	if err != nil {
		return "", nil
	}
	for _, e := range entries {
		if e.IsDir() && e.Name() == javaVersionString {
			return fullExePath, nil
		}
	}

	logger.L.Warn.Printf("[!] Java %d not found! Downloading it now...\n", javaVersion)
	err = installJava(javaVersion)
	if err != nil {
		logger.L.Error.Printf("[!] An error occurred while downloading Java version %d\n", javaVersion)
		logger.L.Info.Println(err)
		return "", err
	}

	if !config.C.Application.Quiet {
		logger.L.Ok.Println("[+] Done!")
	}

	return fullExePath, nil
}
