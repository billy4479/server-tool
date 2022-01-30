package git

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
	"github.com/billy4479/server-tool/logger"
	"github.com/billy4479/server-tool/tui"
)

const (
	gitForWindowsURL = "https://api.github.com/repos/git-for-windows/git/releases"
)

var (
	ErrGitNotInstalled = errors.New("Git was not installed")
)

// TODO: test this
func promptGitInstall() error {
	if runtime.GOOS == "linux" {
		logger.L.Error.Println("[!] To run this server you need Git but it is not installed.")
		logger.L.Error.Println("[!] Please install the version provided by your distribution.")
		return ErrGitNotInstalled
	}

	logger.L.Info.Println("[?] To run this server you need Git. Do you want to install it now?")
	opt, err := tui.MakeMenu(true,
		tui.Option{
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
		tui.Option{
			Description: "No, abort",
			Action:      func() error { return ErrGitNotInstalled },
		},
	)
	if err != nil {
		return err
	}

	if err = opt.Action(); err != nil {
		logger.L.Error.Println("[!] An error has occurred while installing Git:")
		fmt.Printf("\t%s\n", err.Error())
	} else {
		logger.L.Ok.Println("[+] Git installed successfully")
	}

	return err
}
