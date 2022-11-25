package servertool

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"

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
		L.Info.Printf("[?] %s: ", desc)
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
			L.Info.Printf("[?] Your option [1-%d]: ", len(options))
		} else {
			L.Info.Printf("[?] Your option [0-%d] (default: 0): ", len(options)-1)
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
					if !C.Application.Quiet {
						L.Warn.Printf("[!] Option %d was not found.\n", inputN)
					}
					continue
				}
				if !C.Application.Quiet {
					L.Warn.Printf("[!] Option %d was not found, falling back on default.\n", inputN)
				}
				return &options[0], nil
			}

			if !C.Application.Quiet {
				L.Ok.Printf("[+] Option %d selected.\n", inputN)
			}
			return &options[n], nil
		} else if noDefault {
			L.Warn.Println("[!] Invalid option.")
		}
		run = noDefault
	}

	if !C.Application.Quiet {
		L.Ok.Printf("[+] Default option selected.\n")
	}
	return &options[0], nil
}
