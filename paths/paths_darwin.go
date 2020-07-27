// +build darwin

package paths

import "path/filepath"

func determinePaths() {
	localData = filepath.Join(home, "Library", "Application Support", appFolderReverse)
}
