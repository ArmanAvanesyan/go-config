// env_override demonstrates loading a base config from a file and then
// overriding values with environment variables. Environment variables
// with the APP_ prefix take precedence over file values.
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/parser/yaml"
	"github.com/ArmanAvanesyan/go-config/providers/source/env"
	"github.com/ArmanAvanesyan/go-config/providers/source/file"
)

// AppConfig is the example application config (server section, overridable by APP_* env).
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

	// Sources are applied in order: file first, env last (env wins).
	err = config.New().
		AddSource(file.New("config.yaml"), yp).
		AddSource(env.New("APP")).
		Load(ctx, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", cfg)
}
