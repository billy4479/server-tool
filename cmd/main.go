package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"

	st "github.com/billy4479/server-tool"
	"github.com/fatih/color"
	"github.com/skratchdot/open-golang/open"
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

	err := st.LoadConfig()
	if err != nil {
		st.L.Warn.Println("[!] An error has occurred while loading the config file. Falling back on the default...")
		if err = st.WriteConfig(); err != nil {
			return err
		}
	} else {
		st.L.Ok.Println("[+] Config loaded successfully")
	}

	needRestart, err := st.AmITheUpdate(os.Args)
	if err != nil {
		return err
	}
	if needRestart {
		st.L.Ok.Println("[+] Update was successful, restart the application.")
		return nil
	}

	// FIXME: check before forcing the update
	err = st.DoUpdateIfNeeded()
	if err != nil {
		st.L.Error.Println(err)
		st.L.Warn.Println("[!] Unable to update! Proceeding anyways...")

		// We don't crash here
		err = nil
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

	st.L.Ok.Println("[?] What do we do?")
	opt, err := MakeMenu(false,
		Option{
			Description: "Start a server",
			Action: func() error {
				servers, err := st.FindServers()
				if err != nil {
					return err
				}

				st.L.Info.Println("[?] The following servers have been found:")
				c, err := MakeMenu(true, MakeServersMenuItem(servers)...)
				if err != nil {
					return err
				}

				return c.Action()
			},
		},
		Option{
			Description: "Create new a server",
			Action: func() error {
				versions, err := st.GetVersionInfos()
				if err != nil {
					return err
				}

				s := st.Server{}
				s.Name, err = StringOption("Enter a name for the new server", nil)
				if err != nil {
					return err
				}

				s.BaseDir = path.Join(st.C.Application.WorkingDir, s.Name)

				versionStr, err := StringOption(
					"Enter a version for the new server (? to list all versions)",
					func(s string) bool {
						if s == "" {
							return false
						}

						if s == "?" {
							for _, v := range versions {
								fmt.Printf("[+] %s\n", v.ID)
							}
							return false
						}

						for _, v := range versions {
							if v.ID == s {
								return true
							}
						}

						st.L.Warn.Printf("[!] Version %s was not found. Type ? for a list of the available versions\n", s)
						return false
					},
				)

				if err != nil {
					return err
				}

				for _, v := range versions {
					if v.ID == versionStr {
						s.Version = &v
						break
					}
				}

				if s.Version == nil {
					panic("NOT REACHED")
				}

				err = st.CreateServer(&s)
				if err != nil {
					return err
				}
				st.L.Ok.Println("[+] Server created successfully!")
				return nil
			},
		},
		Option{
			Description: "Open server folder",
			Action: func() error {
				return open.Start(st.C.Application.WorkingDir)
			},
		},
		Option{
			Description: "Open config",
			Action: func() error {
				configPath, _, err := st.GetConfigPath()
				if err != nil {
					return err
				}
				return open.Start(configPath)
			},
		},
		Option{
			Description: "Open cache folder",
			Action: func() error {
				return open.Start(st.C.Application.CacheDir)
			},
		},
		Option{
			Description: "Quit",
			Action: func() error {
				return nil
			},
		},
	)
	if err != nil {
		return err
	}
	if err = opt.Action(); err != nil {
		return err
	}

	err = st.WriteConfig()
	if err != nil {
		return err
	}

	return nil
}
