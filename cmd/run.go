package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/billy4479/server-tool/lib"
	"github.com/fatih/color"
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

type UIMode int

const (
	GUI UIMode = iota
	TUI
	CLI
)

func Run() error {
	uiMode := GUI

	defer func() {
		if runtime.GOOS == "windows" && uiMode == TUI {
			fmt.Print("Press enter to continue...")
			fmt.Scanln()
		}
	}()

	color.New(color.FgBlue, color.Bold).Println("[*] Server-Tool version", lib.Version)

	fmt.Printf("[+] OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH)

	if (runtime.GOOS != "windows" &&
		runtime.GOOS != "linux") ||
		runtime.GOARCH != "amd64" {
		return fmt.Errorf("[!] Your OS is not supported!")
	}

	needRestart, err := lib.AmITheUpdate(os.Args)
	if err != nil {
		return err
	}
	if needRestart {
		lib.L.Ok.Println("[+] Update was successful, restart the application.")
		return nil
	}

	err = lib.LoadConfig()
	if err != nil {
		lib.L.Warn.Println("[!] An error has occurred while loading the config file. Falling back on the default...")
		if err = lib.WriteConfig(); err != nil {
			return err
		}
	} else {
		lib.L.Ok.Println("[+] Config loaded successfully")
	}

	if len(os.Args) != 1 {
		if os.Args[1] == "tui" {
			uiMode = TUI
		} else {
			uiMode = CLI
		}
	} else if lib.C.Application.ForceTUI {
		uiMode = TUI
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

	switch uiMode {
	case TUI:
		lib.L.Info.Println("[+] Running in TUI mode\n")
		return runTui()
	case GUI:
		lib.L.Info.Println("[+] Running in GUI mode\n")
		return runGui()
	case CLI:
		lib.L.Info.Println("[+] Running in CLI mode\n")
		return runCli()
	}

	return nil
}
