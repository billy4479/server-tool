package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	st "github.com/billy4479/server-tool"
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
		st.L.Info.Printf("[?] %s: ", desc)
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
			st.L.Info.Printf("[?] Your option [1-%d]: ", len(options))
		} else {
			st.L.Info.Printf("[?] Your option [0-%d] (default: 0): ", len(options)-1)
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
					st.L.Warn.Printf("[!] Option %d was not found.\n", inputN)
					continue
				}
				st.L.Warn.Printf("[!] Option %d was not found, falling back on default.\n", inputN)
				return &options[0], nil
			}

			return &options[n], nil
		} else if noDefault {
			st.L.Warn.Println("[!] Invalid option.")
		}
		run = noDefault
	}

	st.L.Ok.Printf("[+] Default option selected.\n")
	return &options[0], nil
}

func makeServersMenuItem(servers []st.Server) []Option {
	result := []Option{}

	for _, s := range servers {
		desc := fmt.Sprintf("\"%s\" (", s.Name)
		if s.Version == nil {
			desc += "?? on ??"
		} else {
			desc += fmt.Sprintf("%s on ", s.Version.ID)
			switch s.Type {
			case st.Vanilla:
				desc += "Vanilla"
			case st.Fabric:
				desc += "Fabric"
			}
		}

		if s.HasGit {
			desc += " - Git"
		}

		desc += ")"

		result = append(result, Option{
			Description: desc,
			Action:      s.Start,
		})
	}

	return result
}

func runTui() error {
	needUpdate, newVersionURL, err := st.CheckUpdates()
	if err != nil {
		return err
	}

	if needUpdate {
		err = st.DoUpdate(newVersionURL)
		if err != nil {
			st.L.Error.Println(err)
			st.L.Warn.Println("[!] Unable to update! Proceeding anyways...")

			// We don't crash here
			// err = nil
		}
	}

	st.L.Ok.Println("[?] What do we do?")
	opt, err := makeMenu(false,
		Option{
			Description: "Start a server",
			Action: func() error {
				servers, err := st.FindServers()
				if err != nil {
					return err
				}

				st.L.Info.Println("[?] The following servers have been found:")
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
