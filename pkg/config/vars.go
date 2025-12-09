package config

import "path/filepath"

func AppDir(homeDir string) string {
	return filepath.Join(homeDir, ".local", "share", "harvester")
}

func LogDir(homeDir string) string {
	return filepath.Join(AppDir(homeDir), "logs")
}

func LogPath(homeDir string) string {
	return filepath.Join(LogDir(homeDir), "harvester.logs")
}
