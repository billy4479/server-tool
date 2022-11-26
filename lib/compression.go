package lib

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// https://snyk.io/research/zip-slip-vulnerability#go
func checkIllegalPath(dest, name string) error {
	// Here p is already cleaned by filepath.Join
	p := filepath.Join(dest, name)
	expectedPrefix := filepath.Clean(dest) + string(os.PathSeparator)
	if !strings.HasPrefix(p, expectedPrefix) {
		return fmt.Errorf("%s: illegal file path. Expected %s", p, expectedPrefix)
	}
	return nil
}

func moveOutOfSingleFolder(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	if len(entries) != 1 || !entries[0].IsDir() {
		return nil
	}
	singleDirPath := filepath.Join(path, entries[0].Name())
	entries, err = os.ReadDir(singleDirPath)
	if err != nil {
		return err
	}
	for _, e := range entries {
		err = os.Rename(filepath.Join(singleDirPath, e.Name()), filepath.Join(path, e.Name()))
		if err != nil {
			return err
		}
	}

	return os.RemoveAll(singleDirPath)
}

// Thanks to https://stackoverflow.com/questions/28249782/is-it-possible-to-extract-a-tar-xz-package-in-golang
func Untargz(input io.Reader, dest string, skipName string) error {

	// Create an gz Reader
	r, err := gzip.NewReader(input)
	if err != nil {
		return err
	}
	// Create a tar Reader
	tr := tar.NewReader(r)
	basename := ""
	// Iterate through the files in the archive.
	for {
		header, err := tr.Next()
		if err == io.EOF {
			// end of tar archive
			break
		}
		if err != nil {
			return err
		}

		if basename == "" &&
			strings.TrimRight(header.Name, "/") == skipName {
			basename = header.Name
		}

		destPath := filepath.Join(dest, strings.TrimPrefix(header.Name, basename))
		if err = checkIllegalPath(dest, header.Name); err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			// create a directory
			err = os.MkdirAll(destPath, fs.FileMode(header.Mode))
			if err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			// write a file
			file, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(file, tr)
			if err != nil {
				return err
			}
		}
	}

	return moveOutOfSingleFolder(dest)
}

// Thanks to https://golangcode.com/unzip-files-in-go/
func Unzip(input io.ReaderAt, size int64, dest string, skipName string) error {
	r, err := zip.NewReader(input, size)
	if err != nil {
		return err
	}

	basename := ""
	for _, f := range r.File {

		if basename == "" &&
			strings.TrimRight(f.Name, "/") == skipName {
			basename = f.Name
		}

		destPath := filepath.Join(dest, strings.TrimPrefix(f.Name, basename))
		if err = checkIllegalPath(dest, f.Name); err != nil {
			return err
		}

		if f.FileInfo().IsDir() {
			// Make Folder
			err := os.MkdirAll(destPath, os.ModePerm)
			if err != nil {
				return err
			}
			continue
		}

		// Make File
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}
		defer outFile.Close()

		zipFileReader, err := f.Open()
		if err != nil {
			return err
		}
		defer zipFileReader.Close()

		_, err = io.Copy(outFile, zipFileReader)
		if err != nil {
			return err
		}
	}

	return moveOutOfSingleFolder(dest)
}
