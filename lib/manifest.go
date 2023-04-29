package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
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

type VersionType uint8

const (
	VersionTypeRelease VersionType = iota
	VersionTypeSnapshot
	VersionTypeOldBeta
	VersionTypeOldAlpha
)

type VersionInfo struct {
	ID          string
	Type        VersionType
	ReleaseDate time.Time
	JarURL      string
	JavaVersion int
	SHA         string
}

const (
	versionManifestURL = "https://launchermeta.mojang.com/mc/game/version_manifest.json"
	expireTime         = 24 * time.Hour
)

func ManifestPath() string { return filepath.Join(C.Application.CacheDir, "manifest.json") }

func versionTypeStringToEnum(versionType string) (VersionType, error) {
	switch versionType {
	case "old_alpha":
		return VersionTypeOldAlpha, nil
	case "old_beta":
		return VersionTypeOldBeta, nil
	case "release":
		return VersionTypeRelease, nil
	case "snapshot":
		return VersionTypeSnapshot, nil
	default:
		return 0, fmt.Errorf("Invalid version type %s", versionType)
	}
}

type ManifestDownloadProgress interface {
	SetTotal(int)
	Add(string)
	Done()
}

func updateVersionInfos(progress ManifestDownloadProgress) ([]VersionInfo, error) {
	res, err := http.Get(versionManifestURL)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	manifest := VersionManifestJSON{}
	err = json.NewDecoder(res.Body).Decode(&manifest)
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

	progress.SetTotal(len(manifest.Versions))

	for _, v := range manifest.Versions {
		infos.wg.Add(1)
		go func(id, url, versionTypeStr, releaseTimeStr string) {
			defer infos.wg.Done()

			res, err := http.Get(url)
			if err != nil {
				infos.Lock()
				infos.err = err

				L.Error.Printf("[!] An error has occurred while downloading %s: %v\n", id, err)

				infos.Unlock()
				return
			}
			defer res.Body.Close()

			j, err := gabs.ParseJSONBuffer(res.Body)
			if err != nil {
				infos.Lock()
				infos.err = err

				L.Error.Printf("[!] An error has occurred while parsing %s: %v\n", id, infos.err)

				infos.Unlock()
				return
			}
			dl := j.Search("downloads", "server")

			jarURL, okJar := dl.Search("url").Data().(string)
			sha, okSha := dl.Search("sha1").Data().(string)
			javaVersion, okVersion := j.Search("javaVersion", "majorVersion").Data().(float64)

			progress.Add(id)

			if !okJar {
				// Not all of the old versions had multiplayer builtin
				return
			}

			if !okSha {
				L.Warn.Printf("[?] Version %s has no SHA1. Skipping...\n", id)
				return
			}

			if !okVersion {
				// Releases before 1.7 don't have a Java version specified
				// but Java 8 seams to work fine
				javaVersion = 8
			}

			versionType, err := versionTypeStringToEnum(versionTypeStr)
			if err != nil {
				return
			}

			releaseTime, err := time.Parse(time.RFC3339, releaseTimeStr)
			if err != nil {
				return
			}

			info := VersionInfo{
				ID:          id,
				Type:        versionType,
				ReleaseDate: releaseTime,
				JarURL:      jarURL,
				SHA:         sha,
				JavaVersion: int(javaVersion),
			}

			infos.Lock()
			infos.data = append(infos.data, info)
			infos.Unlock()
		}(v.ID, v.URL, v.Type, v.ReleaseTime)
	}

	infos.wg.Wait()
	if infos.err != nil {
		return nil, infos.err
	}

	manifestFile, err := os.Create(ManifestPath())
	if err != nil {
		return nil, err
	}
	defer manifestFile.Close()

	sort.Sort(VerInfos(infos.data))

	encoder := json.NewEncoder(manifestFile)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(infos.data)
	if err != nil {
		return nil, err
	}

	progress.Done()

	return infos.data, nil
}

func GetVersionInfos(progress ManifestDownloadProgress) ([]VersionInfo, error) {
	manifestStat, err := os.Stat(ManifestPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			L.Info.Println("[+] Version manifests are missing. Dowloading them again...")

			return updateVersionInfos(progress)
		}
		L.Error.Printf("[!] Cannot stat %s", ManifestPath())
		return nil, err
	}

	if manifestStat.ModTime().Add(expireTime).Before(time.Now()) {
		L.Info.Println("[+] Version manifests are expired. Dowloading them again...")
		return updateVersionInfos(progress)
	}

	manifestFile, err := os.Open(ManifestPath())
	if err != nil {
		return nil, err
	}
	defer manifestFile.Close()

	versionInfos := []VersionInfo{}
	err = json.NewDecoder(manifestFile).Decode(&versionInfos)
	if err != nil {
		L.Info.Println("[+] Version manifests are corrupted. Dowloading them again...")
		return updateVersionInfos(progress)
	}
	return versionInfos, nil
}

type VerInfos []VersionInfo

func (v VerInfos) Len() int           { return len(v) }
func (v VerInfos) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }
func (v VerInfos) Less(i, j int) bool { return v[i].ReleaseDate.Before(v[j].ReleaseDate) }

func GetVersionInfosSorted(progress ManifestDownloadProgress) ([]VersionInfo, error) {
	vers, err := GetVersionInfos(progress)
	if err != nil {
		return nil, err
	}

	sort.Sort(sort.Reverse(VerInfos(vers)))
	return vers, nil
}
