package servertool

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Jeffail/gabs/v2"
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

const (
	gitForWindowsURL = "https://api.github.com/repos/git-for-windows/git/releases"
)

var (
	ErrGitNotInstalled = errors.New("Git was not installed")
)

// TODO: test this
func promptGitInstall() error {
	if runtime.GOOS == "linux" {
		L.Error.Println("[!] To run this server you need Git but it is not installed.")
		L.Error.Println("[!] Please install the version provided by your distribution.")
		return ErrGitNotInstalled
	}

	L.Info.Println("[?] To run this server you need Git. Do you want to install it now?")
	opt, err := MakeMenu(true,
		Option{
			Description: "Yes, install it now",
			Action: func() error {
				res, err := http.Get(gitForWindowsURL)
				if err != nil {
					return err
				}
				defer res.Body.Close()

				j, err := gabs.ParseJSONBuffer(res.Body)
				if err != nil {
					return err
				}

				for _, rel := range j.Children()[0].Search("assets").Children() {
					if strings.HasSuffix(rel.Search("name").Data().(string), "64-bit.exe") {
						gitInstaller, err := http.Get(rel.Search("browser_download_url").Data().(string))
						if err != nil {
							return err
						}
						defer gitInstaller.Body.Close()

						tmp, err := os.CreateTemp("", "")
						if err != nil {
							return err
						}
						defer func() {
							err := tmp.Truncate(0)
							if err != nil {
								panic(err)
							}
							tmp.Close()
						}()

						_, err = io.Copy(tmp, gitInstaller.Body)
						if err != nil {
							return err
						}

						err = tmp.Sync()
						if err != nil {
							return err
						}

						return exec.Command(tmp.Name()).Run()
					}
				}

				return errors.New("No suitable Git version was found")
			},
		},
		Option{
			Description: "No, abort",
			Action:      func() error { return ErrGitNotInstalled },
		},
	)
	if err != nil {
		return err
	}

	if err = opt.Action(); err != nil {
		L.Error.Println("[!] An error has occurred while installing Git:")
		fmt.Printf("\t%s\n", err.Error())
	} else {
		L.Ok.Println("[+] Git installed successfully")
	}

	return err
}

const (
	lockFileName = "__lock"
)

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
		err = promptGitInstall()
		if err != nil {
			return err
		}
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
		err = promptGitInstall()
		if err != nil {
			return err
		}
	}

	if C.Git.UseLockFile {
		_, err = RunCmdPretty(false, true, baseDir, false, "git", "rm", lockFileName)
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
