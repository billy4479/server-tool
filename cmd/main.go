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
	exitCode := Run()

	if runtime.GOOS == "windows" {
		fmt.Print("Press enter to continue...")
		fmt.Scanln()
	}

	os.Exit(exitCode)
}

func Run() int {
	color.New(color.FgBlue, color.Bold).Println("[*] Server-Tool version", st.Version)

	fmt.Printf("[+] OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH)

	if (runtime.GOOS != "windows" &&
		runtime.GOOS != "linux") ||
		runtime.GOARCH != "amd64" {
		st.L.Error.Println("[!] Your OS is not supported!")
		return 1
	}

	err := st.LoadConfig()
	if err != nil {
		st.L.Warn.Println("[!] An error has occurred while loading the config file. Falling back on the default...")
		if err = st.WriteConfig(); err != nil {
			st.L.Error.Printf("[!] %s\n", err.Error())
			return 1
		}
	} else if !st.C.Application.Quiet {
		st.L.Ok.Println("[+] Config loaded successfully")
	}

	needRestart, err := st.AmITheUpdate(os.Args)
	if err != nil {
		st.L.Error.Println("[!] An error occurred while updating")
		st.L.Debug.Println(err)
		return 1
	}
	if needRestart {
		st.L.Ok.Println("[+] Update was successful, restart the application.")
		st.L.Debug.Println(err)
		return 0
	}

	err = st.Update()
	if err != nil {
		st.L.Error.Println("[!] Unable to update!")
		st.L.Debug.Println(err)

		// We don't crash here
		err = nil
	}

	if err := makeCacheDir(); err != nil {
		st.L.Error.Println("[!] Cache directory cannot be accessed or were not found!")
		fmt.Println(err)
		return 1
	}

	if !st.C.Git.Disable {
		gitVersion, err := st.DetectGit()
		if err != nil {
			st.L.Warn.Println("[!] Git not detected!")
		} else if !st.C.Application.Quiet {
			st.L.Info.Printf("[+] Found Git %s", gitVersion)
		}
	}

	st.L.Ok.Println("[?] What do we do?")
	opt, err := st.MakeMenu(false,
		st.Option{
			Description: "Start a server",
			Action: func() error {
				servers, err := st.FindServers()
				if err != nil {
					return err
				}

				st.L.Info.Println("[?] The following servers have been found:")
				c, err := st.MakeMenu(true, st.MakeServersMenuItem(servers)...)
				if err != nil {
					return err
				}

				return c.Action()
			},
		},
		st.Option{
			Description: "Create new a server",
			Action: func() error {
				versions, err := st.GetVersionInfos()
				if err != nil {
					return err
				}

				s := st.Server{}
				s.Name, err = st.StringOption("Enter a name for the new server", nil)
				if err != nil {
					return err
				}

				s.BaseDir = path.Join(st.C.Application.WorkingDir, s.Name)

				versionStr, err := st.StringOption(
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
		st.Option{
			Description: "Open server folder",
			Action: func() error {
				return open.Start(st.C.Application.WorkingDir)
			},
		},
		st.Option{
			Description: "Open config",
			Action: func() error {
				configPath, _, err := st.GetConfigPath()
				if err != nil {
					return err
				}
				return open.Start(configPath)
			},
		},
		st.Option{
			Description: "Open cache folder",
			Action: func() error {
				return open.Start(st.C.Application.CacheDir)
			},
		},
		st.Option{
			Description: "Quit",
			Action: func() error {
				return nil
			},
		},
	)
	if err != nil {
		st.L.Error.Printf("[!] %s\n", err.Error())
		return 1
	}
	if err = opt.Action(); err != nil {
		st.L.Error.Printf("[!] %s\n", err.Error())
		return 1
	}

	err = st.WriteConfig()
	if err != nil {
		st.L.Error.Printf("[!] Error writing the config %s\n", err.Error())
		return 1
	}

	return 0
}
