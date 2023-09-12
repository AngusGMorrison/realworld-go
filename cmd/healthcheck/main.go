package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Healthcheck failed: %s", err)
	}
}

func run() error {
	port, ok := os.LookupEnv("REALWORLD_PORT")
	if !ok {
		return errors.New("REALWORLD_PORT not set")
	}

	client := http.Client{Timeout: 1 * time.Second}
	healthcheckURL := fmt.Sprintf("http://localhost:%s/healthcheck", port)
	res, err := client.Get(healthcheckURL) //nolint:noctx
	if err != nil {
		return fmt.Errorf("healthcheck client error: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("app server returned status %d", res.StatusCode)
	}

	return nil
}
