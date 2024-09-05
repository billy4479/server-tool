package lib

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)

const ProgName = "server-tool"

func MakeCacheDir() (err error) {
	if C.Application.CacheDir == "" {
		C.Application.CacheDir, err = os.UserCacheDir()
		if err != nil {
			return err
		}
		C.Application.CacheDir =
			filepath.Join(C.Application.CacheDir, ProgName)
	}
	if err = os.MkdirAll(C.Application.CacheDir, 0700); err != nil {
		return err
	}

	return nil
}

func RunCmdPretty(workDir string, name string, args ...string) error {

	cmdLine := name
	if filepath.IsAbs(name) {
		_, cmdLine = path.Split(name)
	}

	for _, arg := range args {
		cmdLine += " " + arg
	}

	L.Debug.Printf("Running \"%s\"\n", cmdLine)

	cmd := exec.Command(name, args...)
	cmd.Stdout = L.Writer
	cmd.Stderr = L.Writer
	cmd.Stdin = os.Stdin
	cmd.Dir = workDir
	addSysProcAttr(cmd)

	L.Info.Println("---!--- Start of command output ---!---")
	err := cmd.Run()
	L.Info.Println("---!--- End of command output ---!---")

	if err != nil {
		if cmd.ProcessState == nil {
			return err
		}
		if !cmd.ProcessState.Success() {
			L.Warn.Printf("Command exited with code %d\n", cmd.ProcessState.ExitCode())
			return fmt.Errorf("Command failed with code %d. The command that failed was %s", cmd.ProcessState.ExitCode(), cmdLine)
		}
	}

	L.Ok.Println("Command exited with code 0")
	return nil
}
