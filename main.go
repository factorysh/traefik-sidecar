package main

import (
	"context"
	"net/http"
	"os"

	"github.com/docker/docker/client"
	"github.com/factorysh/docker-visitor/visitor"
	"github.com/factorysh/traefik-sidecar/events"
	"github.com/factorysh/traefik-sidecar/projects"
	"github.com/factorysh/traefik-sidecar/web"
)

func main() {
	docker, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	w := visitor.New(docker)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Start(ctx)
	p := projects.New(w)
	c, err := events.New("http://localhost:8080", "admin", os.Getenv("PASSWORD"), p)
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	ctx2 := context.Background()
	mux.Handle("/events", web.New(ctx2, c.Events))
	go c.WatchBackends()
	http.ListenAndServe(":3000", mux)

}
