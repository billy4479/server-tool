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
