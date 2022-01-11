package java

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/Jeffail/gabs/v2"
	"github.com/billy4479/server-tool/compression"
	"github.com/billy4479/server-tool/logger"
	"github.com/dustin/go-humanize"
)

func installJava(javaVersion int) error {
	res, err := http.Get(fmt.Sprintf(adoptiumApiUrl, javaVersion))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	j, err := gabs.ParseJSONBuffer(res.Body)
	if err != nil {
		return err
	}

	url := ""
	name := ""
	relName := ""
	var size uint64 = 0
	var checksum []byte = nil

	for _, openjdk := range j.Children() {
		binary := openjdk.Search("binary")

		if binary.Search("os").Data().(string) == runtime.GOOS &&
			binary.Search("architecture").Data().(string) == "x64" &&
			binary.Search("image_type").Data().(string) == "jdk" {

			pack := binary.Search("package")
			url = pack.Search("link").Data().(string)
			name = pack.Search("name").Data().(string)
			size = uint64(pack.Search("size").Data().(float64))
			relName = openjdk.Search("release_name").Data().(string)
			checksum, err = hex.DecodeString(pack.Search("checksum").Data().(string))
			if err != nil {
				return err
			}

			break
		}
	}

	if url == "" || name == "" || checksum == nil || relName == "" {
		return errors.New("Unable to find needed variables in JSON response")
	}

	logger.L.Info.Printf("[+] Downloading %s (%s)\n", name, humanize.Bytes(size))

	res, err = http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	tmp, err := os.CreateTemp("", "")
	if err != nil {
		return err
	}
	defer func() {
		if err := tmp.Truncate(0); err != nil {
			panic(err)
		}
		tmp.Close()
	}()

	_, err = io.Copy(tmp, res.Body)
	if err != nil {
		return err
	}

	err = tmp.Sync()
	if err != nil {
		return err
	}

	{
		_, err = tmp.Seek(0, 0)
		if err != nil {
			return err
		}

		sha := sha256.New()
		_, err = io.Copy(sha, tmp)
		if err != nil {
			return err
		}

		if !bytes.Equal(sha.Sum(nil), checksum) {
			return errors.New("Checksum verification failed")
		}

		_, err = tmp.Seek(0, 0)
		if err != nil {
			return err
		}
	}

	dest := path.Join(javaDir(), fmt.Sprint(javaVersion))

	logger.L.Info.Printf("[+] Extracting %s\n", name)
	// Windows uses .zip, the rest .tar.gz
	if runtime.GOOS == "windows" {
		err = compression.Unzip(tmp, res.ContentLength, dest, relName)
	} else {
		err = compression.Untargz(tmp, dest, relName)
	}

	return err
}
