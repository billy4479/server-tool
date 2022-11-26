package lib

import (
	"errors"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

const ProgName = "server-tool"

var (
	ErrAborted = errors.New("Aborted due to a failed command")
)

func RunCmdPretty(verbose bool, must bool, workDir string, noOutput bool, name string, args ...string) (bool, error) {
	{
		cmdLine := name
		if filepath.IsAbs(name) {
			_, cmdLine = path.Split(name)
		}

		for _, arg := range args {
			cmdLine += " " + arg
		}

		L.Info.Printf("[+] Running \"%s\"\n", cmdLine)
	}
	cmd := exec.Command(name, args...)
	if !noOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	}
	cmd.Stdin = os.Stdin
	cmd.Dir = workDir

	if verbose {
		L.Info.Println("[+] Start of command output")
	}
	err := cmd.Run()
	if verbose {
		L.Info.Println("[+] End of command output")
	}

	if err != nil {
		if cmd.ProcessState == nil {
			return false, err
		}
		if !cmd.ProcessState.Success() {
			if must {
				return false, ErrAborted
			}
			return false, nil
		}
	}

	if verbose {
		L.Ok.Println("[+] Process has exited successfully")
	}

	return true, nil
}
