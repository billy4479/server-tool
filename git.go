package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Jeffail/gabs/v2"
)

var (
	hasGit = false

	ErrGitNotInstalled = errors.New("Git was not installed")
)

const (
	gitForWindowsURL = "https://api.github.com/repos/git-for-windows/git/releases"
)

func detectGit() (string, error) {
	cmd := exec.Command("git", "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	hasGit = true
	return strings.Split(string(out), " ")[2], nil
}

// TODO: test this
func promptGitInstall() error {
	if runtime.GOOS == "linux" {
		Error.Println("[!] To run this server you need Git but it is not installed.")
		Error.Println("[!] Please install the version provided by your distribution.")
		return ErrGitNotInstalled
	}

	Info.Println("[?] To run this server you need Git. Do you want to install it now?")
	opt, err := makeMenu(true,
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

						fmt.Println(tmp.Name())
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
		Error.Println("[!] An error has occurred while installing Git:")
		fmt.Printf("\t%s\n", err.Error())
	} else {
		Ok.Println("[+] Git installed successfully")
	}

	return err
}

func gitPreFn(s *Server) (err error) {
	if config.Git.Disable {
		return nil
	}

	if config.Git.Overrides.Enable {
		pre := config.Git.Overrides.CustomPreCommands
		if len(pre) > 0 {
			for _, cmd := range pre {
				_, err = runCmdPretty(false, true, s.BaseDir, false, cmd[0], cmd[1:]...)
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
	_, err = runCmdPretty(false, true, s.BaseDir, false, "git", "pull")
	return err
}

func gitPostFn(s *Server) (err error) {

	if config.Git.Overrides.Enable {
		post := config.Git.Overrides.CustomPostCommands
		if len(post) > 0 {
			for _, cmd := range post {
				_, err = runCmdPretty(false, true, s.BaseDir, false, cmd[0], cmd[1:]...)
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

	_, err = runCmdPretty(false, true, s.BaseDir, false, "git", "add", "-A")
	if err != nil {
		return err
	}

	_, err = runCmdPretty(false, true, s.BaseDir, false, "git", "commit", "--allow-empty-message", "-m", "")
	if err != nil {
		return err
	}

	_, err = runCmdPretty(false, true, s.BaseDir, false, "git", "push")
	return err
}
