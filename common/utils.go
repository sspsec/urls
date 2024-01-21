package common

import (
	"log"
	"os"
	"path/filepath"
)

func GetConfigFilePath() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	ConfigPath, err := filepath.EvalSymlinks(filepath.Dir(exePath))

	return ConfigPath
}
