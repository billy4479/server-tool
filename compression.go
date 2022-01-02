package main

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

// Thanks to https://golangcode.com/unzip-files-in-go/
func unzip(input io.ReaderAt, size int64, dest string) error {
	r, err := zip.NewReader(input, size)
	if err != nil {
		return err
	}

	basename := ""
	for _, f := range r.File {

		if f.FileInfo().IsDir() &&
			strings.HasPrefix(f.Name, "jdk-") &&
			basename == "" {

			basename = f.Name
		}

		// https://snyk.io/research/zip-slip-vulnerability#go
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
	return nil
}

// Thanks to https://stackoverflow.com/questions/28249782/is-it-possible-to-extract-a-tar-xz-package-in-golang
func untargz(input io.Reader, dest string) error {

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

		if header.Typeflag == tar.TypeDir &&
			strings.Count(header.Name, "/") == 1 &&
			strings.HasPrefix(header.Name, "jdk-") &&
			basename == "" {

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

	return nil
}
