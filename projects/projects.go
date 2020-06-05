package projects

import (
	"errors"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/factorysh/docker-visitor/visitor"
	"github.com/factorysh/traefik-sidecar/traefik"
	log "github.com/sirupsen/logrus"
)

type Projects struct {
	watcher *visitor.Watcher
}

func New(watcher *visitor.Watcher) *Projects {
	return &Projects{
		watcher: watcher,
	}
}

var normalized = regexp.MustCompile(`[^a-zA-Z0-9]`)

// NormalizeName normalizes name, just like traefik does.
func NormalizeName(name string) string {
	return normalized.ReplaceAllString(name, "-")
}

func cutServerName(name string) (string, error) {
	l := strings.LastIndexByte(name, '-')
	if l < 0 { // It should not append
		return "", errors.New("Can't find - in " + name)
	}
	s := strings.IndexByte(name, '-')
	if s < 0 { // It should not append
		return "", errors.New("Can't find - in " + name)
	}
	return NormalizeName(name[s+1 : l]), nil
}

func (p *Projects) projectOfContainer(id string) (string, error) {
	c := p.watcher.Container(id)
	if c == nil {
		return "", errors.New("container not found : " + id)
	}
	return c.Config.Labels["com.docker.compose.project"], nil
}

// Project of a backend
func (p *Projects) Project(b *traefik.Backend) (string, error) {
	p.watcher.Ready()
	for server := range b.Servers {
		name, err := cutServerName(server)
		if err != nil {
			return "", err
		}
		cs, err := p.watcher.Find(func(c *types.ContainerJSON) (bool, error) {
			return NormalizeName(c.Name)[1:] == name, nil
		})
		l := log.WithField("cs", cs)
		if err != nil {
			l.WithError(err).Error("")
			return "", err
		}
		if len(cs) == 0 {
			return "", nil
		}
		if len(cs) > 1 {
			return "", errors.New("Oups, not unique container : " + name)
		}
		// All containers of a backend must be in the same project
		l.WithField("labels", cs[0].Config.Labels).Info("projects.Projects")
		return cs[0].Config.Labels["com.docker.compose.project"], nil
	}
	return "", nil
}
