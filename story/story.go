package story

import (
	"context"
	"encoding/json"
	"os"

	"github.com/factorysh/pubsub/event"
	"github.com/factorysh/traefik-sidecar/metrics"
	log "github.com/sirupsen/logrus"
)

type Story struct {
	events *event.Events
	logger *log.Logger
	path   string
}

func NewStoryWithPath(events *event.Events, path string) *Story {
	return &Story{
		events: events,
		logger: log.New(),
		path:   path,
	}
}

func New(events *event.Events) *Story {
	return &Story{
		events: events,
	}
}

func (s *Story) Listen(ctx context.Context) error {
	if s.logger != nil {
		s.logger.SetFormatter(&log.JSONFormatter{})
		file, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		s.logger.Out = file
		s.logger.Info("start logger")
		defer s.logger.Info("stop logger")
	}
	evts := s.events.Subscribe(ctx)
	for {
		select {
		case evt := <-evts:
			var events map[string]interface{}
			err := json.Unmarshal([]byte(evt.Data), &events)
			if err != nil {
				log.WithError(err).WithField("evt", evt).Error()
				continue
			}
			for k, v := range events {
				var action string
				if v == nil {
					action = "stop"
					metrics.BackendDeath.Inc()
				} else {
					action = "start"
					metrics.BackendBirth.Inc()
				}
				if s.logger != nil {
					s.logger.WithField("backend", k).Info(action)
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}
