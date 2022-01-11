package git

import (
	"github.com/billy4479/server-tool/config"
	"github.com/billy4479/server-tool/utils"
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
	return err
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
