package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/parser/yaml"
	envresolver "github.com/ArmanAvanesyan/go-config/providers/resolver/env"
	"github.com/ArmanAvanesyan/go-config/providers/source/file"
)

// AppConfig is the example application config (server and app sections).
type AppConfig struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
	App struct {
		Name string `json:"name"`
	} `json:"app"`
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
		config.WithResolver(envresolver.New()),
	).AddSource(file.New("examples/basic/config.yaml"), yp).
		Load(ctx, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", cfg)
}
