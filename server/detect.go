package server

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"io"
	"os"

	"github.com/billy4479/server-tool/manifest"
)

func detectServerVersion(serverJarPath string, s *Server) error {
	infos, err := manifest.GetVersionInfos()
	if err != nil {
		return err
	}

	jar, err := os.Open(serverJarPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer jar.Close()

	hasher := sha1.New()
	_, err = io.Copy(hasher, jar)
	if err != nil {
		return err
	}
	sha := hex.EncodeToString(hasher.Sum(nil))

	for _, v := range infos {
		if v.SHA == sha {
			s.Version = &v
			return nil
		}
	}

	return nil
}
