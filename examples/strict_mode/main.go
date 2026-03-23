// strict_mode demonstrates loading config with the strict JSON decoder,
// which rejects any keys in the config file that do not map to a field
// in the target struct.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/decoder/strict"
	"github.com/ArmanAvanesyan/go-config/providers/parser/yaml"
	"github.com/ArmanAvanesyan/go-config/providers/source/file"
)

// AppConfig is the example config for strict_mode (server section).
type AppConfig struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
}

func main() {
	ctx := context.Background()

	yp, err := yaml.New(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = yp.Close(ctx) }()

	var cfg AppConfig

	err = config.New(
		config.WithDecoder(strict.New()),
	).AddSource(file.New("config.yaml"), yp).
		Load(ctx, &cfg)
	if err != nil {
		// In strict mode, unknown fields cause Load to return an error.
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", cfg)
}
