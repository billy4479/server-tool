package compression

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

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

	return nil
}
