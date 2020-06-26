package story

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/factorysh/pubsub/event"
	log "github.com/sirupsen/logrus"
)

type Story struct {
	events *event.Events
	logger *log.Logger
}

func New(events *event.Events) *Story {
	return &Story{
		events: events,
		logger: log.New(),
	}
}

func (s *Story) Listen(ctx context.Context) error {
	evts := s.events.Subscribe(ctx)
	for {
		select {
		case evt := <-evts:
			fmt.Println(evt)
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
				} else {
					action = "start"
				}
				s.logger.WithField("backend", k).Info(action)
			}
		case <-ctx.Done():
			return nil
		}
	}
}
