package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

var (
	configOnce sync.Once
	APIBaseURL = "http://localhost:8080"
)

func initConfig() {
	configOnce.Do(func() {
		loadDotEnv()
		if value := strings.TrimRight(strings.TrimSpace(os.Getenv("API_BASE_URL")), "/"); value != "" {
			APIBaseURL = value
		}
	})
}

func loadDotEnv() {
	workingDir, err := os.Getwd()
	if err != nil {
		return
	}

	for directory := workingDir; ; directory = filepath.Dir(directory) {
		_ = godotenv.Load(filepath.Join(directory, ".env"))
		if directory == filepath.Dir(directory) {
			break
		}
	}
}

func apiURL(route string) string {
	initConfig()
	return APIBaseURL + route
}
