package compression

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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
	return nil
}
