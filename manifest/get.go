package manifest

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/billy4479/server-tool/config"
	"github.com/billy4479/server-tool/logger"
)

func GetVersionInfos() ([]VersionInfo, error) {
	err := os.MkdirAll(manifestDir(), 0700)
	if err != nil {
		return nil, err
	}
	manifestStat, err := os.Stat(versionManifest())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if config.C.Application.Quiet {
				logger.L.Info.Println("[+] Updating manifests")
			} else {
				logger.L.Info.Println("[+] Version manifests are missing. Dowloading them again...")
			}

			return updateVersionInfos()
		}
		logger.L.Error.Printf("[!] Cannot stat %s", versionManifest())
		return nil, err
	}

	if manifestStat.ModTime().Add(expireTime).Before(time.Now()) {
		if config.C.Application.Quiet {
			logger.L.Info.Println("[+] Updating manifests")
		} else {
			logger.L.Info.Println("[+] Version manifests are expired. Dowloading them again...")
		}
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
		logger.L.Info.Println("[+] Version manifests are corrupted. Dowloading them again...")
		return updateVersionInfos()
	}
	return versionInfos, nil
}
