package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/fatih/color"
)

func main() {
	color.New(color.FgBlue, color.Bold).Println("[*] Server-Tool")

	fmt.Printf("OS: %s, Arch: %s\n", runtime.GOOS, runtime.GOARCH)

	if (runtime.GOOS != "windows" &&
		runtime.GOOS != "linux") ||
		runtime.GOARCH != "amd64" {
		Error.Println("[!] Your OS is not supported!")
		os.Exit(1)
		return
	}

	baseDir := os.Getenv("BASE_DIR")
	if baseDir != "" {
		Info.Printf("[+] Using %s as work directory\n", baseDir)
	}

	c, err := makeChoiceMenu(false,
		Choice{
			Description: "Start a server",
			Action: func() error {
				servers, err := findServers()
				if err != nil {
					return err
				}

				Info.Println("The following servers have been found:")
				c, err := makeChoiceMenu(true, makeServersMenuItem(servers)...)
				if err != nil {
					return err
				}

				return c.Action()
			},
		},
		Choice{
			Description: "Create new a server",
			Action: func() error {
				Error.Println("[!] TODO")
				return nil
			},
		},
	)
	if err != nil {
		Error.Println(err)
		os.Exit(1)
	}
	if err = c.Action(); err != nil {
		Error.Println(err)
		os.Exit(1)
	}
}
