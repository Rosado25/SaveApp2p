package helpers

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

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

func jwtSecret() string {
	loadDotEnv()
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if secret == "" {
		panic("JWT_SECRET is not configured")
	}
	return secret
}
