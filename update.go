package servertool

import (
	"errors"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"
)

func DoUpdateIfNeeded() error {
	needUpdate, newVersionURL, err := CheckUpdates()
	if err != nil {
		return err
	}

	if needUpdate {
		L.Info.Println("[+] A new version as been found!")
		if err != nil {
			return err
		}

		err = do(newVersionURL)
		if err != nil {
			return err
		}

		// Done, just use the new one
		os.Exit(2)
	}

	return nil
}

func AmITheUpdate(args []string) (bool, error) {
	if len(args) == 3 {
		if args[1] == "replace" {
			exe, err := os.Executable()
			if err != nil {
				return true, err
			}

			me, err := os.Open(exe)
			if err != nil {
				return true, err
			}
			defer me.Close()

			old, err := os.Create(args[2])
			if err != nil {
				return true, err
			}
			defer old.Close()

			_, err = io.Copy(old, me)
			if err != nil {
				return true, err
			}

			// config.C.Application.WorkingDir = filepath.Dir(args[2])
			return true, nil
		}
	}

	return false, nil
}

func do(URL string) error {
	L.Info.Printf("[+] Downloading %s\n", URL)
	new, err := http.Get(URL)
	if err != nil {
		return err
	}
	defer new.Body.Close()

	tmp, err := os.CreateTemp("", "*.exe")
	if err != nil {
		return err
	}

	_, err = io.Copy(tmp, new.Body)
	if err != nil {
		tmp.Close()
		return err
	}

	err = tmp.Chmod(0700)
	tmp.Close()
	if err != nil {
		return err
	}

	L.Ok.Println("[+] Done. Updating now")

	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(tmp.Name(), "replace", exe)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	return cmd.Start()

}

var (
	Version = "dev"
)

const (
	releaseURL = "https://api.github.com/repos/billy4479/server-tool/releases"
)

func CheckUpdates() (bool, string, error) {
	if Version == "dev" {
		L.Info.Println("[+] This is a development build, skipping updates.")
		return false, "", nil
	}

	res, err := http.Get(releaseURL)
	if err != nil {
		return false, "", err
	}
	defer res.Body.Close()

	j, err := gabs.ParseJSONBuffer(res.Body)
	if err != nil {
		return false, "", err
	}

	tagName := j.Children()[0].Search("tag_name").Data().(string)
	for _, rel := range j.Children()[0].Search("assets").Children() {
		if strings.Contains(rel.Search("name").Data().(string), runtime.GOOS) {
			downloadURL := rel.Search("browser_download_url").Data().(string)

			s := strings.Split(strings.TrimLeft(Version, "v"), "-")

			current, err := strconv.ParseInt(strings.ReplaceAll(s[0], ".", ""), 10, 32)
			if err != nil {
				return false, "", err
			}

			remote, err := strconv.ParseInt(strings.ReplaceAll(strings.TrimLeft(tagName, "v"), ".", ""), 10, 32)
			if err != nil {
				return false, "", err
			}

			if remote > current {
				return true, downloadURL, nil
			}

			// return true, downloadURL, nil
			return false, "", nil
		}
	}

	return false, "", errors.New("Parsing Error")
}
