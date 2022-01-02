package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/fatih/color"
)

type option struct {
	Description string
	Action      func() error
	PrintFn     func(position int, noDefault bool)
}

var ErrNotEnoughoptions = errors.New("At least one option is required")

func makeoptionMenu(noDefault bool, options ...option) (*option, error) {
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
			Info.Printf("Your option [1-%d]: ", len(options))
		} else {
			Info.Printf("Your option [0-%d] (default: 0): ", len(options)-1)
		}
		input := ""
		fmt.Scanln(&input)
		if inputN, err := strconv.ParseInt(input, 10, 32); err == nil {
			n := int(inputN)
			if noDefault {
				n -= 1
			}

			if n >= len(options) || n < 0 {
				if noDefault {
					Warn.Printf("option %d was not found.\n", inputN)
					continue
				}
				Warn.Printf("option %d was not found, falling back on default.\n", inputN)
				return &options[0], nil
			}

			Ok.Printf("option %d selected.\n", inputN)
			return &options[n], nil
		} else if noDefault {
			Warn.Println("Invalid option.")
		}
		run = noDefault
	}
	Ok.Printf("Default option selected.\n")
	return &options[0], nil
}
