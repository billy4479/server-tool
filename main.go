package main

import (
	"os"

	"github.com/billy4479/server-tool/cmd"
	"github.com/billy4479/server-tool/lib"
)

func main() {
	err := cmd.Run()
	if err != nil {
		lib.L.Error.Printf("[!] %s", err)
		os.Exit(1)
	}
}
