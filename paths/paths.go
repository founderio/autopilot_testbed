// Package paths supports determination of system and user folders for configuration, application data, caches, etc.
// Heavily based on these two MIT licensed libraries:
// * https://github.com/kirsle/configdir
// * https://github.com/shibukawa/configdir
package paths

import (
	"log"

	"github.com/mitchellh/go-homedir"
)

const (
	appFolder        = "autopilot_testbed"
	appFolderReverse = "net.founderio.autopilot_testbed"
)

var (
	localData string
	home      string

	pathsDetermined bool
)

func ensurePathsDetermined() {
	if !pathsDetermined {
		var err error
		home, err = homedir.Dir()
		if err != nil {
			log.Fatalln("Unable to determine user home directory:", err.Error())
		}

		determinePaths()
		pathsDetermined = true
	}
}

func GetDataPath() string {
	ensurePathsDetermined()
	return localData
}
