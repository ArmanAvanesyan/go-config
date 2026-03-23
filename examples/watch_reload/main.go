// watch_reload demonstrates live-reload with WatchTyped, a file watcher, and config diff.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/parser/json"
	"github.com/ArmanAvanesyan/go-config/providers/source/file"
	"github.com/ArmanAvanesyan/go-config/runtime/diff"
	"github.com/ArmanAvanesyan/go-config/runtime/watch/fsnotify"
)

// AppConfig is the example config struct used for watch reload.
type AppConfig struct {
	Server struct {
		Port int `json:"port"`
	} `json:"server"`
}

func main() {
	dir := os.TempDir()
	configPath := filepath.Join(dir, "watch_reload_config.json")
	if err := os.WriteFile(configPath, []byte(`{"server":{"port":8080}}`), 0644); err != nil {
		log.Fatal(err)
	}

	loader := config.New().
		AddSource(file.New(configPath), json.New())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Trigger a file change after a short delay to demonstrate reload + diff
	go func() {
		time.Sleep(200 * time.Millisecond)
		_ = os.WriteFile(configPath, []byte(`{"server":{"port":9090}}`), 0644)
	}()
	go func() {
		time.Sleep(2 * time.Second)
		cancel()
	}()

	err := config.WatchTyped[AppConfig](ctx, loader, fsnotify.New(configPath), func(old, new *AppConfig, changes []diff.Change) {
		if old == nil {
			fmt.Printf("Initial config: server.port=%d\n", new.Server.Port)
			return
		}
		fmt.Printf("Reload: server.port %d -> %d\n", old.Server.Port, new.Server.Port)
		for _, c := range changes {
			fmt.Printf("  diff %s: %v -> %v\n", c.Path, c.OldValue, c.NewValue)
		}
	})
	if err != nil && err != context.Canceled {
		log.Fatal(err)
	}
	fmt.Println("Done.")
}
