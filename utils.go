package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const progName = "server-tool"

var (
	ErrAborted = errors.New("Aborted due to a failed command")

	inputReader = bufio.NewReader(os.Stdin)
)

func populateDataDirs() (err error) {
	if config.Application.CacheDir == "" {
		config.Application.CacheDir, err = os.UserCacheDir()
		if err != nil {
			return err
		}
		config.Application.CacheDir = filepath.Join(config.Application.CacheDir, progName)
	}
	if err = os.MkdirAll(config.Application.CacheDir, 0700); err != nil {
		return err
	}

	return nil
}

func makeServersMenuItem(servers []Server) []Option {
	result := []Option{}

	for _, s := range servers {
		desc := fmt.Sprintf("\"%s\" (", s.Name)
		if s.Version == nil {
			desc += "?? on ??"
		} else {
			desc += fmt.Sprintf("%s on ", s.Version.ID)
			switch s.Type {
			case Vanilla:
				desc += "Vanilla"
			case Fabric:
				desc += "Fabric"
			}
		}

		if s.HasGit {
			desc += " - Git"
		}
		if s.HasStartScript {
			desc += ", Start Script"
		}

		desc += ")"

		result = append(result, Option{
			Description: desc,
			Action:      s.Start,
		})
	}

	return result
}

func runCmdPretty(verbose bool, must bool, workDir string, noOutput bool, name string, args ...string) (bool, error) {
	{
		cmdLine := name
		if filepath.IsAbs(name) {
			_, cmdLine = path.Split(name)
		}

		for _, arg := range args {
			cmdLine += " " + arg
		}

		Info.Printf("[+] Running \"%s\"\n", cmdLine)
	}
	cmd := exec.Command(name, args...)
	if !noOutput {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stdout
	}
	cmd.Stdin = os.Stdin
	cmd.Dir = workDir

	if verbose {
		Info.Println("[+] Start of command output")
	}
	err := cmd.Run()
	if verbose {
		Info.Println("[+] End of command output")
	}

	if err != nil {
		if cmd.ProcessState == nil {
			return false, err
		}
		if !cmd.ProcessState.Success() {
			Error.Printf("[!] Process has terminated with error code %d\n", cmd.ProcessState.ExitCode())
			if must {
				return false, ErrAborted
			}
			return false, nil
		}
	}

	if verbose {
		Ok.Println("[+] Process has exited successfully")
	}

	return true, nil
}

func readLine() (string, error) {
	input, err := inputReader.ReadString('\n')
	if err != nil {
		return "", err
	}
	input = strings.ReplaceAll(input, "\n", "")
	return input, nil
}
