package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/billy4479/server-tool/config"
	"github.com/billy4479/server-tool/utils"
)

const (
	lockFileName = "__lock"
)

func PreFn(baseDir string) (err error) {
	if config.C.Git.Disable {
		return nil
	}

	if config.C.Git.Overrides.Enable {
		pre := config.C.Git.Overrides.CustomPreCommands
		if len(pre) > 0 {
			for _, cmd := range pre {
				_, err = utils.RunCmdPretty(false, true, baseDir, false, cmd[0], cmd[1:]...)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	if !hasGit {
		err = promptGitInstall()
		if err != nil {
			return err
		}
	}
	_, err = utils.RunCmdPretty(false, true, baseDir, false, "git", "pull")
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
		if config.C.Git.UseLockFile {
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

			_, err = utils.RunCmdPretty(false, true, baseDir, false, "git", "add", "-A")
			if err != nil {
				return err
			}

			_, err = utils.RunCmdPretty(false, true, baseDir, false, "git", "commit", "-m", "Pushing lock file")
			if err != nil {
				return err
			}

			_, err = utils.RunCmdPretty(false, true, baseDir, false, "git", "push")
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

	if config.C.Git.Overrides.Enable {
		post := config.C.Git.Overrides.CustomPostCommands
		if len(post) > 0 {
			for _, cmd := range post {
				_, err = utils.RunCmdPretty(false, true, baseDir, false, cmd[0], cmd[1:]...)
				if err != nil {
					return err
				}
			}
			return nil
		}
	}

	if !hasGit {
		err = promptGitInstall()
		if err != nil {
			return err
		}
	}

	if config.C.Git.UseLockFile {
		_, err = utils.RunCmdPretty(false, true, baseDir, false, "git", "rm", lockFileName)
		if err != nil {
			return err
		}
	}

	_, err = utils.RunCmdPretty(false, true, baseDir, false, "git", "add", "-A")
	if err != nil {
		return err
	}

	_, err = utils.RunCmdPretty(false, true, baseDir, false, "git", "commit", "--allow-empty-message", "-m", "")
	if err != nil {
		return err
	}

	_, err = utils.RunCmdPretty(false, true, baseDir, false, "git", "push")
	return err
}
