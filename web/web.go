package web

import (
	"context"
	"net/http"

	"github.com/factorysh/pubsub/event"
	"github.com/factorysh/pubsub/sse"
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

func (w *Web) ServeHTTP(wr http.ResponseWriter, r *http.Request) {
	sse.HandleSSE(w.ctx, w.events, wr, log.WithField("beuha", "aussi"), 0)
}
