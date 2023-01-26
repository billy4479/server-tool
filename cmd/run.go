package cmd

import (
	"fmt"
	"runtime"

	"github.com/billy4479/server-tool/lib"
)

func Run() error {
	if (runtime.GOOS != "windows" &&
		runtime.GOOS != "linux") ||
		runtime.GOARCH != "amd64" {
		return fmt.Errorf("[!] Your OS is not supported!")
	}

	err := lib.LoadConfig()
	if err != nil {
		// lib.L.Warn.Println("[!] An error has occurred while loading the config file. Falling back on the default...")
		if err = lib.WriteConfig(); err != nil {
			return err
		}
	}

	if err := lib.SetupLogger(); err != nil {
		return err
	}
	defer lib.L.Close()

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

	err = runCli()
	if err != nil {
		lib.L.Error.Printf("[!] FATAL ERROR: %v", err)
	}
	return err
}
