package main

import (
	"context"
	"net/http"
	"os"

	"github.com/factorysh/traefik-sidecar/events"
	"github.com/factorysh/traefik-sidecar/web"
)

func main() {
	c, err := events.New("http://localhost:8080", "admin", os.Getenv("PASSWORD"))
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()
	ctx := context.Background()
	mux.Handle("/events", web.New(ctx, c.Events))
	go c.WatchBackends()
	http.ListenAndServe(":3000", mux)

}
