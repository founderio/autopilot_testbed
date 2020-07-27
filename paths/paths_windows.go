// +build windows

package paths

import (
	"os"
	"path/filepath"
)

func determinePaths() {
	localData = filepath.Join(os.Getenv("APPDATA"), appFolder)
}
