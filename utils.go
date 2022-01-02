package main

import "os"

func getWorkDir() string {
	baseDir := os.Getenv("BASE_DIR")
	if baseDir != "" {
		return baseDir
	}

	return "."
}
