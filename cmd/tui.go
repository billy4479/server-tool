package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	st "github.com/billy4479/server-tool"
	"github.com/fatih/color"
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

func MakeMenu(noDefault bool, options ...Option) (*Option, error) {
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

func MakeServersMenuItem(servers []st.Server) []Option {
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
