package compression

import (
	"fmt"
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
