package lib

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
	"github.com/dustin/go-humanize"
)

func javaExeName() string {
	if runtime.GOOS == "windows" {
		return "java.exe"
	}
	return "java"
}

const adoptiumApiUrl = "https://api.adoptium.net/v3/assets/latest/%d/hotspot?os=%s&architecture=x64&image_type=jre"

func JavaDir() string     { return path.Join(C.Application.CacheDir, "java") }
func javaExePath() string { return path.Join("bin", javaExeName()) }

func installJava(javaVersion int) error {
	res, err := http.Get(fmt.Sprintf(adoptiumApiUrl, javaVersion, runtime.GOOS))
	if err != nil {
		return err
	}
	defer res.Body.Close()

	j, err := gabs.ParseJSONBuffer(res.Body)
	if err != nil {
		return err
	}

	openjdk := j.Children()[0]

	binary := openjdk.Search("binary")
	pack := binary.Search("package")
	url := pack.Search("link").Data().(string)
	name := pack.Search("name").Data().(string)
	size := uint64(pack.Search("size").Data().(float64))
	relName := openjdk.Search("release_name").Data().(string)
	checksum, err := hex.DecodeString(pack.Search("checksum").Data().(string))
	if err != nil {
		return err
	}

	if url == "" || name == "" || checksum == nil || relName == "" {
		return errors.New("Unable to find needed variables in JSON response")
	}

	L.Info.Printf("[+] Downloading %s (%s)\n", name, humanize.Bytes(size))

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

	dest := path.Join(JavaDir(), fmt.Sprint(javaVersion))

	L.Info.Printf("[+] Extracting %s\n", name)
	// Windows uses .zip, the rest .tar.gz
	if runtime.GOOS == "windows" {
		err = Unzip(tmp, res.ContentLength, dest, relName)
	} else {
		err = Untargz(tmp, dest, relName)
	}

	return err
}

func EnsureJavaIsInstalled(javaVersion int) (string, error) {
	javaVersionString := fmt.Sprint(javaVersion)
	err := os.MkdirAll(JavaDir(), 0700)
	if err != nil {
		return "", nil
	}

	fullExePath := path.Join(JavaDir(), javaVersionString, javaExePath())

	entries, err := os.ReadDir(JavaDir())
	if err != nil {
		return "", nil
	}
	for _, e := range entries {
		if e.IsDir() && e.Name() == javaVersionString {
			return fullExePath, nil
		}
	}

	L.Warn.Printf("[!] Java %d not found! Downloading it now...\n", javaVersion)
	err = installJava(javaVersion)
	if err != nil {
		L.Error.Printf("[!] An error occurred while downloading Java version %d\n", javaVersion)
		L.Info.Println(err)
		return "", err
	}

	L.Ok.Println("[+] Done!")

	return fullExePath, nil
}
