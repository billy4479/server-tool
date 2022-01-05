package main

import (
	"fmt"
	"os"
	"path"
	"runtime"

	"github.com/fatih/color"
)

func main() {
	color.New(color.FgBlue, color.Bold).Println("[*] Server-Tool")

	fmt.Printf("[+] OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH)

	if (runtime.GOOS != "windows" &&
		runtime.GOOS != "linux") ||
		runtime.GOARCH != "amd64" {
		Error.Println("[!] Your OS is not supported!")
		os.Exit(1)
		return
	}

	err := loadConfig()
	if err != nil {
		Warn.Println("[!] An error has occurred while loading the config file. Falling back on the default...")
		if err = writeConfig(); err != nil {
			Error.Printf("[!] %s\n", err.Error())
			os.Exit(1)

		}
	} else if !config.Application.Quiet {
		Ok.Println("[+] Config loaded successfully")
	}

	if err := populateDataDirs(); err != nil {
		Error.Println("[!] Data directories cannot be accessed or were not found!")
		fmt.Println(err)
		os.Exit(1)
		return
	}

	if !config.Git.Disable {
		gitVersion, err := detectGit()
		if err != nil {
			Warn.Println("[!] Git not detected!")
		} else if !config.Application.Quiet {
			Info.Printf("[+] Found Git %s", gitVersion)
		}
	}

	Ok.Println("[?] What do we do?")
	opt, err := makeMenu(false,
		Option{
			Description: "Start a server",
			Action: func() error {
				servers, err := findServers()
				if err != nil {
					return err
				}

				Info.Println("[?] The following servers have been found:")
				c, err := makeMenu(true, makeServersMenuItem(servers)...)
				if err != nil {
					return err
				}

				return c.Action()
			},
		},
		Option{
			Description: "Create new a server",
			Action: func() error {
				versions, err := getVersionInfos()
				if err != nil {
					return err
				}

				s := Server{}
				for s.Name == "" {
					Info.Print("[?] Enter a name for the new server: ")
					s.Name, err = readLine()
					if err != nil {
						return err
					}

					s.BaseDir = path.Join(config.Application.WorkingDir, s.Name)
				}

				for s.Version == nil {
					Info.Print("[?] Enter a version for the new server (? to list all versions): ")
					versionStr, err := readLine()
					if err != nil {
						return err
					}

					if versionStr == "?" {
						for _, v := range versions {
							fmt.Printf("[+] %s\n", v.ID)
						}
						continue
					}

					for _, v := range versions {
						if v.ID == versionStr {
							s.Version = &v
							break
						}
					}

					if s.Version == nil {
						Warn.Println("[!] The chosen version was not found! Type ? for a list of the available versions")
					}
				}

				err = createServer(&s)
				if err != nil {
					return err
				}
				Ok.Println("[+] Server created successfully!")
				return nil
			},
		},
	)
	if err != nil {
		Error.Printf("[!] %s\n", err.Error())
		os.Exit(1)
	}
	if err = opt.Action(); err != nil {
		Error.Printf("[!] %s\n", err.Error())
		os.Exit(1)
	}
}
