package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	st "github.com/billy4479/server-tool"
	"github.com/fatih/color"
)

func makeCacheDir() (err error) {
	if st.C.Application.CacheDir == "" {
		st.C.Application.CacheDir, err = os.UserCacheDir()
		if err != nil {
			return err
		}
		st.C.Application.CacheDir =
			filepath.Join(st.C.Application.CacheDir, st.ProgName)
	}
	if err = os.MkdirAll(st.C.Application.CacheDir, 0700); err != nil {
		return err
	}

	return nil
}

func main() {
	err := Run()
	if err != nil {
		st.L.Error.Printf("[!] %s", err)
		os.Exit(1)
	}
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

	color.New(color.FgBlue, color.Bold).Println("[*] Server-Tool version", st.Version)

	fmt.Printf("[+] OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH)

	if (runtime.GOOS != "windows" &&
		runtime.GOOS != "linux") ||
		runtime.GOARCH != "amd64" {
		return fmt.Errorf("[!] Your OS is not supported!")
	}

	needRestart, err := st.AmITheUpdate(os.Args)
	if err != nil {
		return err
	}
	if needRestart {
		st.L.Ok.Println("[+] Update was successful, restart the application.")
		return nil
	}

	err = st.LoadConfig()
	if err != nil {
		st.L.Warn.Println("[!] An error has occurred while loading the config file. Falling back on the default...")
		if err = st.WriteConfig(); err != nil {
			return err
		}
	} else {
		st.L.Ok.Println("[+] Config loaded successfully")
	}

	if len(os.Args) != 1 {
		if os.Args[1] == "tui" {
			uiMode = TUI
		} else {
			uiMode = CLI
		}
	} else if st.C.Application.ForceTUI {
		uiMode = TUI
	}

	if err := makeCacheDir(); err != nil {
		return err
	}

	if !st.C.Git.Disable {
		gitVersion, err := st.DetectGit()
		if err != nil {
			st.L.Warn.Println("[!] Git not detected!")
		} else {
			st.L.Info.Printf("[+] Found Git %s", gitVersion)
		}
	}

	switch uiMode {
	case TUI:
		st.L.Info.Println("[+] Running in TUI mode\n")
		return runTui()
	case GUI:
		st.L.Info.Println("[+] Running in GUI mode\n")
		return runGui()
	case CLI:
		st.L.Info.Println("[+] Running in CLI mode\n")
		return runCli()
	}

	return nil
}
