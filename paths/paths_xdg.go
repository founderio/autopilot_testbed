// +build !windows,!darwin

package paths

import (
	"os"
	"path/filepath"
)

func determinePaths() {
	xdgDataHome := os.Getenv("XDG_DATA_HOME")
	if len(xdgDataHome) > 0 {
		localData = filepath.Join(xdgDataHome, appFolder)
	} else {
		localData = filepath.Join(home, ".local/share", appFolder)
	}
}
