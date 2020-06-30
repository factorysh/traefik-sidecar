package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/factorysh/pubsub/event"
	"github.com/factorysh/pubsub/sse"
	"github.com/factorysh/traefik-sidecar/version"
	log "github.com/sirupsen/logrus"
)

type Web struct {
	ctx    context.Context
	events *event.Events
}

func New(ctx context.Context, events *event.Events) *Web {
	return &Web{
		ctx:    ctx,
		events: events,
	}
}

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, `
    _____
 ___ |[]|_n__n_I_c
|___||__|###|____}
 O-O--O-O+++--O-O

Sidecar %s

Events : /events
`, version.Version())

}

func (w *Web) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	sse.HandleSSE(w.ctx, w.events, wr, log.WithField("beuha", "aussi"), 0)
}
