package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/billy4479/server-tool/lib"
)

func makeCacheDir() (err error) {
	if lib.C.Application.CacheDir == "" {
		lib.C.Application.CacheDir, err = os.UserCacheDir()
		if err != nil {
			return err
		}
		lib.C.Application.CacheDir =
			filepath.Join(lib.C.Application.CacheDir, lib.ProgName)
	}
	if err = os.MkdirAll(lib.C.Application.CacheDir, 0700); err != nil {
		return err
	}

	return nil
}

func Run() error {

	if (runtime.GOOS != "windows" &&
		runtime.GOOS != "linux") ||
		runtime.GOARCH != "amd64" {
		return fmt.Errorf("[!] Your OS is not supported!")
	}

	err := lib.LoadConfig()
	if err != nil {
		lib.L.Warn.Println("[!] An error has occurred while loading the config file. Falling back on the default...")
		if err = lib.WriteConfig(); err != nil {
			return err
		}
	} else {
		lib.L.Ok.Println("[+] Config loaded successfully")
	}

	if err := makeCacheDir(); err != nil {
		return err
	}

	if !lib.C.Git.Disable {
		gitVersion, err := lib.DetectGit()
		if err != nil {
			lib.L.Warn.Println("[!] Git not detected!")
		} else {
			lib.L.Info.Printf("[+] Found Git %s", gitVersion)
		}
	}

	defer func() {
		if err = lib.WriteConfig(); err != nil {
			lib.L.Warn.Printf("Error while saving config: %v", err)
		}
	}()

	return runCli()
}
