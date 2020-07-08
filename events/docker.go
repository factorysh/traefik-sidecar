package events

import (
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/factorysh/docker-visitor/visitor"
	"github.com/factorysh/pubsub/event"
)

func WatchForTraefikContainer(watcher *visitor.Watcher, events *event.Events) error {
	watcher.WatchFor(func(action string, container *types.ContainerJSON) {
		if action == visitor.START || action == visitor.STOP {

			events.Append(&event.Event{
				Event: "docker",
				Id:    events.NextEventId(),
				Data:  fmt.Sprintf(`{"action":"%s", "name":"%s"}`, action, container.Name),
			})
		}
	}, "traefik.frontend.rule")

	return nil
}
