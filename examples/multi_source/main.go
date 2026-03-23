// multi_source demonstrates layering multiple config sources: in-memory
// defaults, a YAML file, and environment variable overrides.
// Each subsequent source overrides the previous one.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/parser/yaml"
	"github.com/ArmanAvanesyan/go-config/providers/source/env"
	"github.com/ArmanAvanesyan/go-config/providers/source/file"
	"github.com/ArmanAvanesyan/go-config/providers/source/memory"
)

// AppConfig is the example application config (server and log, from defaults + file + env).
type AppConfig struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
	Log struct {
		Level string `json:"level"`
	} `json:"log"`
}

func main() {
	ctx := context.Background()

	defaults := memory.New(map[string]any{
		"server": map[string]any{"host": "localhost", "port": 8080},
		"log":    map[string]any{"level": "info"},
	})

	var cfg AppConfig

	yp, err := yaml.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = yp.Close(ctx) }()

	// Precedence (lowest → highest): defaults → file → env
	err = config.New().
		AddSource(defaults).
		AddSource(file.New("config.yaml"), yp).
		AddSource(env.New("APP")).
		Load(ctx, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", cfg)
}
