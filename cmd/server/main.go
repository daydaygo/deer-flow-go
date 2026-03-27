package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/user/deer-flow-go/internal/config"
	"github.com/user/deer-flow-go/internal/handler"
)

func main() {
	configPath := getConfigPath()
	cfg := config.MustLoad(configPath)

	router := handler.NewRouter(cfg)
	mux := http.NewServeMux()
	router.Register(mux)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Starting server on %s\n", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}

func getConfigPath() string {
	if path := os.Getenv("CONFIG_PATH"); path != "" {
		return path
	}
	return "config.yaml"
}
