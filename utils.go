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

var (
	ErrAborted = errors.New("Aborted due to a failed command")

	inputReader = bufio.NewReader(os.Stdin)

	cacheDir  string
	configDir string
)

func populateDataDirs() (err error) {
	cacheDir, err = os.UserCacheDir()
	if err != nil {
		return err
	}
	cacheDir = filepath.Join(cacheDir, "server-tool")
	if err = os.MkdirAll(cacheDir, 0700); err != nil {
		return err
	}

	configDir, err = os.UserConfigDir()
	if err != nil {
		return err
	}
	configDir = filepath.Join(configDir, "server-tool")
	if err = os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	return nil
}

func getWorkDir() string {
	baseDir := os.Getenv("BASE_DIR")
	if baseDir != "" {
		return baseDir
	}

	return "."
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
			case Forge:
				desc += "Forge"
			case Paper:
				desc += "PaperMC"
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

func runCmdPretty(verbose bool, must bool, workDir string, name string, args ...string) (bool, error) {
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
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
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
