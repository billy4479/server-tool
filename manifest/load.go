package manifest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"

	"github.com/Jeffail/gabs/v2"
	"github.com/billy4479/server-tool/config"
	"github.com/billy4479/server-tool/logger"
)

func updateVersionInfos() ([]VersionInfo, error) {
	res, err := http.Get(versionManifestURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	manifestFile, err := os.Create(versionManifest())
	if err != nil {
		return nil, err
	}
	defer manifestFile.Close()

	manifest := VersionManifestJSON{}
	err = json.NewDecoder(io.TeeReader(res.Body, manifestFile)).Decode(&manifest)
	if err != nil {
		return nil, err
	}

	var infos struct {
		data []VersionInfo
		err  error
		wg   sync.WaitGroup
		sync.Mutex
	}

	infos.data = []VersionInfo{}

	for _, v := range manifest.Versions {
		infos.wg.Add(1)
		id := v.ID
		url := v.URL
		go func() {
			defer infos.wg.Done()

			res, err := http.Get(url)
			if err != nil {
				infos.Lock()
				infos.err = err

				logger.L.Error.Printf("[!] An error has occurred while downloading %s\n", id)
				fmt.Println(err)

				infos.Unlock()
				return
			}
			defer res.Body.Close()

			j, err := gabs.ParseJSONBuffer(res.Body)
			if err != nil {
				infos.Lock()
				infos.err = err

				logger.L.Error.Printf("[!] An error has occurred while parsing %s\n", id)
				fmt.Println(infos.err)

				infos.Unlock()
				return
			}
			dl := j.Search("downloads", "server")

			jarURL, okJar := dl.Search("url").Data().(string)
			sha, okSha := dl.Search("sha1").Data().(string)
			javaVersion, okVersion := j.Search("javaVersion", "majorVersion").Data().(float64)

			if !okJar {
				// Not all of the old versions had multiplayer builtin
				return
			}

			if !okSha {
				logger.L.Warn.Printf("[!] Version %s has no SHA1. Skipping...\n", id)
				return
			}

			if !okVersion {
				// Releases before 1.7 don't have a Java version specified
				// but Java 8 seams to work fine
				javaVersion = 8
			}

			info := VersionInfo{
				ID:          id,
				JarURL:      jarURL,
				SHA:         sha,
				JavaVersion: int(javaVersion),
			}

			infos.Lock()
			infos.data = append(infos.data, info)
			infos.Unlock()

			if !config.C.Application.Quiet {
				fmt.Printf("    [+] %s                   \r", id)
			}
		}()
	}

	infos.wg.Wait()
	if infos.err != nil {
		return nil, infos.err
	}

	infosFile, err := os.Create(versionInfos())
	if err != nil {
		return nil, err
	}
	defer infosFile.Close()

	encoder := json.NewEncoder(infosFile)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(infos.data)
	if err != nil {
		return nil, err
	}

	if !config.C.Application.Quiet {
		// TODO: Find a better way...
		logger.L.Ok.Println("[+] Done                      ")
	}

	return infos.data, nil
}
