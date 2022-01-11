package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/billy4479/server-tool/cmd"
)

func main() {
	exitCode := cmd.Run()

	if runtime.GOOS == "windows" {
		fmt.Print("Press enter to continue...")
		fmt.Scanln()
	}

	os.Exit(exitCode)
}
