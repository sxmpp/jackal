// +build windows

package app

import (
	"os"
	"path/filepath"
)

const (
	winEnvKey = "APPDATA"
)

func (a *Application) setRLimit() error {
	return nil
}

func (a *Application) getDefaultConfigDir() string {
	appDataDir := os.Getenv(winEnvKey)

	return filepath.Join(appDataDir, defaultConfigPathPostfix)
}
