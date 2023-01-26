package main

import (
	"os"

	"github.com/billy4479/server-tool/cmd"
)

func main() {
	err := cmd.Run()
	if err != nil {
		os.Exit(1)
	}
}
