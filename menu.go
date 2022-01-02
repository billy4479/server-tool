package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/fatih/color"
)

type Choice struct {
	Description string
	Action      func() error
	PrintFn     func(position int, noDefault bool)
}

var ErrNotEnoughChoices = errors.New("At least one choice is required")

func makeChoiceMenu(noDefault bool, choices ...Choice) (*Choice, error) {
	if len(choices) == 0 {
		return nil, ErrNotEnoughChoices
	}

	run := true
	for run {
		for i, c := range choices {
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
			Info.Printf("Your choice [1-%d]: ", len(choices))
		} else {
			Info.Printf("Your choice [0-%d] (default: 0): ", len(choices)-1)
		}
		input := ""
		fmt.Scanln(&input)
		if inputN, err := strconv.ParseInt(input, 10, 32); err == nil {
			n := int(inputN)
			if noDefault {
				n -= 1
			}

			if n >= len(choices) || n < 0 {
				if noDefault {
					Warn.Printf("Choice %d was not found.\n", inputN)
					continue
				}
				Warn.Printf("Choice %d was not found, falling back on default.\n", inputN)
				return &choices[0], nil
			}

			Ok.Printf("Choice %d selected.\n", inputN)
			return &choices[n], nil
		} else if noDefault {
			Warn.Println("Invalid choice.")
		}
		run = noDefault
	}
	Ok.Printf("Default choice selected.\n")
	return &choices[0], nil
}
