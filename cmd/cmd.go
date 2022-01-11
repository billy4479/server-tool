package cmd

import (
	"fmt"
	"path"
	"runtime"

	"github.com/billy4479/server-tool/config"
	"github.com/billy4479/server-tool/git"
	"github.com/billy4479/server-tool/logger"
	"github.com/billy4479/server-tool/manifest"
	"github.com/billy4479/server-tool/server"
	"github.com/billy4479/server-tool/tui"
	"github.com/fatih/color"
)

func Run() int {
	color.New(color.FgBlue, color.Bold).Println("[*] Server-Tool")

	fmt.Printf("[+] OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH)

	if (runtime.GOOS != "windows" &&
		runtime.GOOS != "linux") ||
		runtime.GOARCH != "amd64" {
		logger.L.Error.Println("[!] Your OS is not supported!")
		return 1
	}

	err := config.LoadConfig()
	if err != nil {
		logger.L.Warn.Println("[!] An error has occurred while loading the config file. Falling back on the default...")
		if err = config.WriteConfig(); err != nil {
			logger.L.Error.Printf("[!] %s\n", err.Error())
			return 1
		}
	} else if !config.C.Application.Quiet {
		logger.L.Ok.Println("[+] Config loaded successfully")
	}

	if err := makeCacheDir(); err != nil {
		logger.L.Error.Println("[!] Cache directory cannot be accessed or were not found!")
		fmt.Println(err)
		return 1
	}

	if !config.C.Git.Disable {
		gitVersion, err := git.DetectGit()
		if err != nil {
			logger.L.Warn.Println("[!] Git not detected!")
		} else if !config.C.Application.Quiet {
			logger.L.Info.Printf("[+] Found Git %s", gitVersion)
		}
	}

	logger.L.Ok.Println("[?] What do we do?")
	opt, err := tui.MakeMenu(false,
		tui.Option{
			Description: "Start a server",
			Action: func() error {
				servers, err := server.FindServers()
				if err != nil {
					return err
				}

				logger.L.Info.Println("[?] The following servers have been found:")
				c, err := tui.MakeMenu(true, server.MakeServersMenuItem(servers)...)
				if err != nil {
					return err
				}

				return c.Action()
			},
		},
		tui.Option{
			Description: "Create new a server",
			Action: func() error {
				versions, err := manifest.GetVersionInfos()
				if err != nil {
					return err
				}

				s := server.Server{}
				s.Name, err = tui.StringOption("Enter a name for the new server", nil)
				if err != nil {
					return err
				}

				s.BaseDir = path.Join(config.C.Application.WorkingDir, s.Name)

				versionStr, err := tui.StringOption(
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

						logger.L.Warn.Printf("[!] Version %s was not found. Type ? for a list of the available versions\n", s)
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

				err = server.CreateServer(&s)
				if err != nil {
					return err
				}
				logger.L.Ok.Println("[+] Server created successfully!")
				return nil
			},
		},
	)
	if err != nil {
		logger.L.Error.Printf("[!] %s\n", err.Error())
		return 1
	}
	if err = opt.Action(); err != nil {
		logger.L.Error.Printf("[!] %s\n", err.Error())
		return 1
	}

	return 0
}
