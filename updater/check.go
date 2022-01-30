package updater

import (
	"errors"
	"net/http"
	"runtime"
	"strconv"
	"strings"

	"github.com/Jeffail/gabs/v2"
)

var (
	Version = "dev"
)

const (
	releaseURL = "https://api.github.com/repos/billy4479/server-tool/releases"
)

func checkUpdates() (bool, string, error) {

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

			// Just in case I need this in the future
			// if len(s) > 1 {
			// This is a development version
			// }

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
