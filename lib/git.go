package lib

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

var (
	hasGit = false
)

func DetectGit() (string, error) {
	cmd := exec.Command("git", "--version")
	addSysProcAttr(cmd)
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
	if !C.Git.Enable {
		return nil
	}

	err := RunCmdPretty(baseDir, "git", "reset", "--hard")
	if err != nil {
		return err
	}

	err = RunCmdPretty(baseDir, "git", "clean", "-fdx")
	if err != nil {
		return err
	}

	remote, err := hasRemotes(baseDir)
	if err != nil {
		return err
	}

	if remote {
		err = RunCmdPretty(baseDir, "git", "fetch", "--all")
		if err != nil {
			return err
		}

		// FIXME: we assume the name of the branches here
		err = RunCmdPretty(baseDir, "git", "reset", "--hard", "origin/master")
		if err != nil {
			return err
		}
	}

	return err
}

func UnfuckCommit(baseDir string) error {
	if !C.Git.Enable {
		return nil
	}

	// Remove lock if present
	err := RunCmdPretty(baseDir, "git", "rm", "-f", "--ignore-unmatch", lockFileName)
	if err != nil {
		return err
	}

	err = RunCmdPretty(baseDir, "git", "add", "-A")
	if err != nil {
		return err
	}

	err = RunCmdPretty(baseDir, "git", "commit", "-m", "Unfuck")
	if err != nil {
		return err
	}

	err = RunCmdPretty(baseDir, "git", "push")
	return err
}

func getGitUsername() (string, error) {
	cmd := exec.Command("git", "config", "user.name")
	name, err := cmd.Output()
	addSysProcAttr(cmd)
	return strings.TrimSpace(string(name)), err
}

func PreFn(baseDir string, progress GitProgress) (err error) {
	if !C.Git.Enable {
		return nil
	}
	if !hasGit {
		return fmt.Errorf("Git not found. Install Git and try again")
	}

	dialog := progress()
	defer dialog("")

	dialog("Checking remotes")
	remotes, err := hasRemotes(baseDir)
	if err != nil {
		return err
	}

	if remotes {
		dialog("Pulling latest changes")
		err = RunCmdPretty(baseDir, "git", "pull")
		if err != nil {
			return err
		}
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
		dialog("Creating lockfile")
		if C.Git.UseLockFile {
			{
				out, err := getGitUsername()
				if err != nil {
					return err
				}

				f, err := os.Create(lockFilePath)
				if err != nil {
					return err
				}
				defer f.Close()

				_, err = f.WriteString(out)
				if err != nil {
					return err
				}
			}

			err = RunCmdPretty(baseDir, "git", "add", "-A")
			if err != nil {
				return err
			}

			err = RunCmdPretty(baseDir, "git", "commit", "-m", "Acquiring lock")
			if err != nil {
				return err
			}

			dialog("Pushing lock file")
			if remotes {
				err = RunCmdPretty(baseDir, "git", "push")
				if err != nil {
					return err
				}
			}
		}
	} else {
		return err
	}

	return nil
}

func hasRemotes(baseDir string) (bool, error) {
	cmd := exec.Command("git", "remote")
	addSysProcAttr(cmd)
	cmd.Dir = baseDir
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}

	remotes := strings.Split(strings.ReplaceAll(string(out), "\r", ""), "\n")
	L.Debug.Printf("Found the following remotes %v (%d)", remotes, len(remotes))

	return len(remotes) != 1, nil
}

func PostFn(baseDir string, progress GitProgress) (err error) {
	if !hasGit {
		return fmt.Errorf("Git not found. Install Git and try again")
	}

	dialog := progress()
	defer dialog("")

	if C.Git.UseLockFile {
		dialog("Removing lock file")
		err = RunCmdPretty(baseDir, "git", "rm", "-f", lockFileName)
		if err != nil {
			return err
		}
	}

	dialog("Adding new files")
	err = RunCmdPretty(baseDir, "git", "add", "-A")
	if err != nil {
		return err
	}

	dialog("Committing files")

	msg := ""
	if serverStartTime != nil {
		msg = fmt.Sprintf("Server started at %s, time played: %s", serverStartTime.Format(time.RFC3339), time.Now().Sub(*serverStartTime).String())
	} else {
		msg = "Unknown server start time"
	}

	err = RunCmdPretty(baseDir, "git", "commit", "-m", msg)
	if err != nil {
		return err
	}

	dialog("Checking remotes")
	remotes, err := hasRemotes(baseDir)
	if err != nil {
		return err
	}

	if remotes {
		dialog("Pushing files")
		err = RunCmdPretty(baseDir, "git", "push")
	}
	return err
}
