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
				javaExePath, err := ensureJavaIsInstalled(16)
				if err != nil {
					return err
				}
				Ok.Printf("[+] Java executable is %s\n", javaExePath)
				return nil
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
