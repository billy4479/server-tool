package main

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
)

func javaExeName() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

var (
	javaDir     = path.Join(getWorkDir(), ".server-tool", "java")
	javaExePath = path.Join("bin", javaExeName())

	adoptiumApiUrl = "https://api.adoptium.net/v3/assets/latest/%d/hotspot?release=latest&jvm_impl=hotspot&vendor=adoptium"
)

func ensureJavaIsInstalled(javaVersion int) (string, error) {
	javaVersionString := fmt.Sprint(javaVersion)
	err := os.MkdirAll(javaDir, 0700)
	if err != nil {
		return "", nil
	}

	fullExePath := path.Join(javaDir, javaVersionString, javaExePath)

	entries, err := os.ReadDir(javaDir)
	if err != nil {
		return "", nil
	}
	for _, e := range entries {
		if e.IsDir() && e.Name() == javaVersionString {
			return fullExePath, nil
		}
	}

	Warn.Printf("[!] Java %d not found! Downloading it now...\n", javaVersion)
	err = installJava(javaVersion)
	if err != nil {
		Error.Printf("An error occurred while downloading Java version %d\n", javaVersion)
		fmt.Println(err)
		return "", err
	}

	Ok.Println("[+] Done")

	return fullExePath, nil
}

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
	var checksum []byte = nil

	for _, openjdk := range j.Children() {
		binary := openjdk.Search("binary")

		// if binary.Search("os").Data().(string) == runtime.GOOS &&
		if binary.Search("os").Data().(string) == runtime.GOOS &&
			binary.Search("architecture").Data().(string) == "x64" &&
			binary.Search("image_type").Data().(string) == "jdk" {

			pack := binary.Search("package")
			url = pack.Search("link").Data().(string)
			name = pack.Search("name").Data().(string)
			checksum, err = hex.DecodeString(pack.Search("checksum").Data().(string))
			if err != nil {
				return err
			}

			break
		}
	}

	if url == "" || name == "" || checksum == nil {
		return errors.New("Unable to find needed variables in JSON response")
	}

	Info.Printf("[+] Downloading %s\n", name)

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

	dest := path.Join(javaDir, fmt.Sprint(javaVersion))

	Info.Printf("[+] Extracting %s\n", name)
	// Windows uses .zip, the rest .tar.gz
	if runtime.GOOS == "windows" {
		err = unzip(tmp, res.ContentLength, dest)
	} else {
		err = untargz(tmp, dest)
	}

	return err
}
