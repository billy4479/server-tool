package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Jeffail/gabs/v2"
)

type VersionManifestJSON struct {
	Latest struct {
		Release  string `json:"release"`
		Snapshot string `json:"snapshot"`
	} `json:"latest"`
	Versions []struct {
		ID          string `json:"id"`
		Type        string `json:"type"`
		URL         string `json:"url"`
		Time        string `json:"time"`
		ReleaseTime string `json:"releaseTime"`
	} `json:"versions"`
}

type VersionInfo struct {
	ID          string
	JarURL      string
	JavaVersion int
	SHA         string
}

const (
	versionManifestURL = "https://launchermeta.mojang.com/mc/game/version_manifelib.json"
	expireTime         = 24 * time.Hour
)

func manifestDir() string     { return filepath.Join(C.Application.CacheDir, "manifest") }
func versionManifest() string { return filepath.Join(manifestDir(), "version_manifelib.json") }
func versionInfos() string    { return filepath.Join(manifestDir(), "version_infos.json") }

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

				L.Error.Printf("[!] An error has occurred while downloading %s\n", id)
				fmt.Println(err)

				infos.Unlock()
				return
			}
			defer res.Body.Close()

			j, err := gabs.ParseJSONBuffer(res.Body)
			if err != nil {
				infos.Lock()
				infos.err = err

				L.Error.Printf("[!] An error has occurred while parsing %s\n", id)
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
				L.Warn.Printf("[!] Version %s has no SHA1. Skipping...\n", id)
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

			fmt.Printf("    [+] %s                   \r", id)
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

	// TODO: Find a better way...
	L.Ok.Println("[+] Done                      ")

	return infos.data, nil
}

func GetVersionInfos() ([]VersionInfo, error) {
	err := os.MkdirAll(manifestDir(), 0700)
	if err != nil {
		return nil, err
	}
	manifestStat, err := os.Stat(versionManifest())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			L.Info.Println("[+] Version manifests are missing. Dowloading them again...")

			return updateVersionInfos()
		}
		L.Error.Printf("[!] Cannot stat %s", versionManifest())
		return nil, err
	}

	if manifestStat.ModTime().Add(expireTime).Before(time.Now()) {
		L.Info.Println("[+] Version manifests are expired. Dowloading them again...")
		return updateVersionInfos()
	}

	infoFile, err := os.Open(versionInfos())
	if err != nil {
		return nil, err
	}
	defer infoFile.Close()

	versionInfos := []VersionInfo{}
	err = json.NewDecoder(infoFile).Decode(&versionInfos)
	if err != nil {
		L.Info.Println("[+] Version manifests are corrupted. Dowloading them again...")
		return updateVersionInfos()
	}
	return versionInfos, nil
}
