package cmd

import (
	"os"
	"path/filepath"

	"github.com/billy4479/server-tool/config"
	"github.com/billy4479/server-tool/utils"
)

func makeCacheDir() (err error) {
	if config.C.Application.CacheDir == "" {
		config.C.Application.CacheDir, err = os.UserCacheDir()
		if err != nil {
			return err
		}
		config.C.Application.CacheDir =
			filepath.Join(config.C.Application.CacheDir, utils.ProgName)
	}
	if err = os.MkdirAll(config.C.Application.CacheDir, 0700); err != nil {
		return err
	}

	return nil
}
