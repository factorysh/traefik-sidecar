package projects

import (
	"errors"

	"github.com/factorysh/docker-visitor/visitor"
)

type Projects struct {
	watcher *visitor.Watcher
}

func New(watcher *visitor.Watcher) *Projects {
	return &Projects{
		watcher: watcher,
	}
}

func (p *Projects) projectOfContainer(id string) (string, error) {
	c := p.watcher.Container(id)
	if c == nil {
		return "", errors.New("container not found : " + id)
	}
	return c.Config.Labels["com.docker.compose.project"], nil
}
