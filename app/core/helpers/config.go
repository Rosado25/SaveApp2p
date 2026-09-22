package helpers

import (
	"log"
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
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		envPath := filepath.Join(exeDir, ".env")

		log.Println("Executável:", exePath)
		log.Println("Procurando .env em:", envPath)

		if _, err := os.Stat(envPath); err == nil {
			log.Println(".env EXISTE")
		} else {
			log.Println(".env NÃO EXISTE:", err)
		}

		if loadDotEnvFrom(exeDir) {
			log.Println("Env carregado do diretório do executável:", exeDir)
			return
		}

		log.Println("Não conseguiu carregar .env do executável")
	}

	if workingDir, err := os.Getwd(); err == nil {
		log.Println("Working directory:", workingDir)

		if loadDotEnvFrom(workingDir) {
			log.Println("Env carregado do working directory")
			return
		}

		log.Println("Não conseguiu carregar .env do working directory")
	}
}

func loadDotEnvFrom(startDir string) bool {
	for directory := startDir; ; directory = filepath.Dir(directory) {
		envPath := filepath.Join(directory, ".env")

		if err := godotenv.Load(envPath); err == nil {
			return true
		}

		if directory == filepath.Dir(directory) {
			break
		}
	}

	return false
}

func apiURL(route string) string {
	initConfig()
	return APIBaseURL + route
}
