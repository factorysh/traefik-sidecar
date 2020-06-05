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
	"github.com/onrik/logrus/filename"
	log "github.com/sirupsen/logrus"
)

func main() {
	filenameHook := filename.NewHook()
	log.AddHook(filenameHook)
	log.SetLevel(log.DebugLevel)
	docker, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	w := visitor.New(docker)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go w.Start(ctx)
	log.Info("Listen docker events")
	p := projects.New(w)
	c, err := events.New("http://localhost:8080", "admin", os.Getenv("PASSWORD"), p)
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	ctx2 := context.Background()
	mux.Handle("/events", web.New(ctx2, c.Events))
	go c.WatchBackends()
	log.Info("watch traefik's backends")
	log.Info("Listening HTTP")
	http.ListenAndServe(":3000", mux)

}
