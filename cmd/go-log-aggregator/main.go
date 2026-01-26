package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"go-log-aggregator/internal/config"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config/config.json", "path to config file")
	flag.Parse()

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if len(cfg.Sources) == 0 {
		fmt.Fprintln(os.Stdout, "no sources configured")
		return
	}

	fmt.Fprintln(os.Stdout, "configured sources:")
	for _, src := range cfg.Sources {
		fmt.Fprintf(os.Stdout, "- %s (%s) format=%s\n", src.Name, src.Path, src.Format)
	}
}
