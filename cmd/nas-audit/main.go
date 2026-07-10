package main

import (
	"flag"
	"fmt"
	"os"

	"nas-audit/internal/config"
)

func main() {
	configPath := flag.String("config", "config/config.example.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("NAS Audit config loaded")
	fmt.Printf("database: %s %s\n", cfg.Database.Driver, cfg.Database.Path)
	fmt.Println("roots:")
	for _, root := range cfg.Roots {
		fmt.Printf("- %s: %s\n", root.Name, root.Path)
	}
}
