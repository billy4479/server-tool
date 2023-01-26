package lib

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	hasGit = false
)

func DetectGit() (string, error) {
	cmd := exec.Command("git", "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	hasGit = true
	return strings.Split(string(out), " ")[2], nil
}

var (
	ErrGitNotInstalled = errors.New("Git was not installed")
)

const (
	lockFileName = "__lock"
)

func UnfuckReset(baseDir string) error {
	if C.Git.Disable {
		return nil
	}

	_, err := RunCmdPretty(false, true, baseDir, false, "git", "reset", "--hard")
	if err != nil {
		return err
	}

	_, err = RunCmdPretty(false, true, baseDir, false, "git", "clean", "-fdx")
	return err
}

func UnfuckCommit(baseDir string) error {
	if C.Git.Disable {
		return nil
	}

	// Remove lock if present
	_, err := RunCmdPretty(false, true, baseDir, false, "git", "rm", "-f", "--ignore-unmatch", lockFileName)
	if err != nil {
		return err
	}

	_, err = RunCmdPretty(false, true, baseDir, false, "git", "add", "-A")
	if err != nil {
		return err
	}

	// We allow this to fail
	success, err := RunCmdPretty(false, false, baseDir, false, "git", "commit", "-m", "Unfuck")
	if err != nil {
		return err
	}

	if success {
		_, err = RunCmdPretty(false, true, baseDir, false, "git", "push")
		if err != nil {
			return err
		}
	}
	return nil
}

func PreFn(baseDir string) (err error) {
	if C.Git.Disable {
		return nil
	}

	if C.Git.Overrides.Enable {
		pre := C.Git.Overrides.CustomPreCommands
		if len(pre) > 0 {
			for _, cmd := range pre {
				_, err = RunCmdPretty(false, true, baseDir, false, cmd[0], cmd[1:]...)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	if !hasGit {
		return fmt.Errorf("Git not found. Install Git and try again")
	}

	_, err = RunCmdPretty(false, true, baseDir, false, "git", "pull")
	if err != nil {
		return err
	}

	lockFilePath := filepath.Join(baseDir, lockFileName)
	if _, err := os.Stat(lockFilePath); err == nil {
		// This fails even if the lock file is disabled, better safe than sorry
		b, err := os.ReadFile(lockFilePath)
		s := string(b)
		if err != nil || len(s) == 0 {
			s = "????"
		}
		return fmt.Errorf("A lockfile was found! The server is probably being used by %s, aborting.", s[:len(s)-1])
	} else if errors.Is(err, os.ErrNotExist) {
		if C.Git.UseLockFile {
			{
				f, err := os.Create(lockFilePath)
				if err != nil {
					return err
				}
				defer f.Close()

				cmd := exec.Command("git", "config", "user.name")
				cmd.Stderr = os.Stderr
				out, err := cmd.Output()
				if err != nil {
					return err
				}

				_, err = f.Write(out)
				if err != nil {
					return err
				}
			}

			_, err = RunCmdPretty(false, true, baseDir, false, "git", "add", "-A")
			if err != nil {
				return err
			}

			_, err = RunCmdPretty(false, true, baseDir, false, "git", "commit", "-m", "Pushing lock file")
			if err != nil {
				return err
			}

			_, err = RunCmdPretty(false, true, baseDir, false, "git", "push")
			if err != nil {
				return err
			}
		}
	} else {
		return err
	}

	return nil
}

func PostFn(baseDir string) (err error) {

	if C.Git.Overrides.Enable {
		post := C.Git.Overrides.CustomPostCommands
		if len(post) > 0 {
			for _, cmd := range post {
				_, err = RunCmdPretty(false, true, baseDir, false, cmd[0], cmd[1:]...)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	if !hasGit {
		return fmt.Errorf("Git not found. Install Git and try again")
	}

	if C.Git.UseLockFile {
		_, err = RunCmdPretty(false, true, baseDir, false, "git", "rm", "-f", lockFileName)
		if err != nil {
			return err
		}
	}

	_, err = RunCmdPretty(false, true, baseDir, false, "git", "add", "-A")
	if err != nil {
		return err
	}

	_, err = RunCmdPretty(false, true, baseDir, false, "git", "commit", "--allow-empty-message", "-m", "")
	if err != nil {
		return err
	}

	_, err = RunCmdPretty(false, true, baseDir, false, "git", "push")
	return err
}
