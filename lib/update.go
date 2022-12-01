package lib

import (
	"errors"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"
	"github.com/minio/selfupdate"
)

func DoUpdate(newVersionURL string) error {
	resp, err := http.Get(newVersionURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = selfupdate.Apply(resp.Body, selfupdate.Options{})
	if err != nil {
		L.Warn.Printf("[!] Update failed: %v\n", err)
		if err = selfupdate.RollbackError(err); err != nil {
			L.Error.Println("[!] Rolling back also failed, you're on your own now")
			return err
		}
		// We return nil here anyways because a failing update shouldn't crash the app
	}
	return nil
}

var (
	Version = "dev"
)

const (
	releaseURL = "https://api.github.com/repos/billy4479/server-tool/releases"
)

func CheckUpdates() (bool, string, error) {

	// Comment to test updates
	if Version == "dev" || strings.Contains(Version, "-") {
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
	assets := j.Children()[0].Search("assets")
	for _, asset := range assets.Children() {
		if strings.Contains(asset.Search("name").Data().(string), runtime.GOOS) {
			downloadURL := asset.Search("browser_download_url").Data().(string)

			s := strings.TrimLeft(Version, "v")
			// s = strings.Split(s, "-")[0] // Uncomment to test updates

			current, err := strconv.ParseInt(strings.ReplaceAll(s, ".", ""), 10, 32)
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

			// return true, downloadURL, nil // Uncomment to test updates
			return false, "", nil
		}
	}

	return false, "", errors.New("Parsing Error")
}
