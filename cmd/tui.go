package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/billy4479/server-tool/lib"
	"github.com/fatih/color"
	"github.com/skratchdot/open-golang/open"
)

var inputReader = bufio.NewReader(os.Stdin)

func readLine() (string, error) {
	input, err := inputReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.ReplaceAll(input, "\r", "")
	input = strings.ReplaceAll(input, "\n", "")
	return input, nil
}

func StringOption(desc string, continueAskingUntil func(string) bool) (string, error) {
	input := ""
	if continueAskingUntil == nil {
		continueAskingUntil = func(s string) bool { return s != "" }
	}

	for !continueAskingUntil(input) {
		lib.L.Info.Printf("[?] %s: ", desc)
		var err error
		input, err = readLine()
		if err != nil {
			return "", err
		}
	}

	return input, nil
}

type Option struct {
	Description string
	Action      func() error
	PrintFn     func(position int, noDefault bool)
}

var ErrNotEnoughoptions = errors.New("At least one option is required")

func makeMenu(noDefault bool, options ...Option) (*Option, error) {
	if len(options) == 0 {
		return nil, ErrNotEnoughoptions
	}

	run := true
	for run {
		for i, c := range options {
			if c.PrintFn != nil {
				c.PrintFn(i, noDefault)
			} else {
				if noDefault {
					color.Cyan("- [%d] %s", i+1, c.Description)
				} else {
					color.Cyan("- [%d] %s", i, c.Description)
				}
			}
		}
		if noDefault {
			lib.L.Info.Printf("[?] Your option [1-%d]: ", len(options))
		} else {
			lib.L.Info.Printf("[?] Your option [0-%d] (default: 0): ", len(options)-1)
		}
		input, err := readLine()
		if err != nil {
			return nil, err
		}

		if inputN, err := strconv.ParseInt(input, 10, 32); err == nil {
			n := int(inputN)
			if noDefault {
				n -= 1
			}

			if n >= len(options) || n < 0 {
				if noDefault {
					lib.L.Warn.Printf("[!] Option %d was not found.\n", inputN)
					continue
				}
				lib.L.Warn.Printf("[!] Option %d was not found, falling back on default.\n", inputN)
				return &options[0], nil
			}

			return &options[n], nil
		} else if noDefault {
			lib.L.Warn.Println("[!] Invalid option.")
		}
		run = noDefault
	}

	lib.L.Ok.Printf("[+] Default option selected.\n")
	return &options[0], nil
}

func makeServersMenuItem(servers []lib.Server) []Option {
	result := []Option{}

	for _, s := range servers {
		desc := fmt.Sprintf("\"%s\" (", s.Name)
		if s.Version == nil {
			desc += "?? on ??"
		} else {
			desc += fmt.Sprintf("%s on ", s.Version.ID)
			switch s.Type {
			case lib.Vanilla:
				desc += "Vanilla"
			case lib.Fabric:
				desc += "Fabric"
			}
		}

		if s.HasGit {
			desc += " - Git"
		}

		desc += ")"

		result = append(result, Option{
			Description: desc,
			Action:      func() error { return s.Start(lib.C.Minecraft.GUI) },
		})
	}

	return result
}

func runTui() error {
	needUpdate, newVersionURL, err := lib.CheckUpdates()
	if err != nil {
		return err
	}

	if needUpdate {
		err = lib.DoUpdate(newVersionURL)
		if err != nil {
			lib.L.Error.Println(err)
			lib.L.Warn.Println("[!] Unable to update! Proceeding anyways...")

			// We don't crash here
			// err = nil
		}
	}

	lib.L.Ok.Println("[?] What do we do?")
	opt, err := makeMenu(false,
		Option{
			Description: "Start a server",
			Action: func() error {
				servers, err := lib.FindServers()
				if err != nil {
					return err
				}

				lib.L.Info.Println("[?] The following servers have been found:")
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
				versions, err := lib.GetVersionInfos()
				if err != nil {
					return err
				}

				s := lib.Server{}
				s.Name, err = StringOption("Enter a name for the new server", nil)
				if err != nil {
					return err
				}

				s.BaseDir = path.Join(lib.C.Application.WorkingDir, s.Name)

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

						lib.L.Warn.Printf("[!] Version %s was not found. Type ? for a list of the available versions\n", s)
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

				err = lib.CreateServer(&s)
				if err != nil {
					return err
				}
				lib.L.Ok.Println("[+] Server created successfully!")
				return nil
			},
		},
		Option{
			Description: "Open server folder",
			Action: func() error {
				return open.Start(lib.C.Application.WorkingDir)
			},
		},
		Option{
			Description: "Open config",
			Action: func() error {
				configPath, _, err := lib.GetConfigPath()
				if err != nil {
					return err
				}
				return open.Start(configPath)
			},
		},
		Option{
			Description: "Open cache folder",
			Action: func() error {
				return open.Start(lib.C.Application.CacheDir)
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

	return nil
}
