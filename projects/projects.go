package projects

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/factorysh/docker-visitor/visitor"
	"github.com/factorysh/traefik-sidecar/traefik"
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

func NormalizeName(name string) string {
	return normalized.ReplaceAllString(name, "-")
}

func (p *Projects) projectOfContainer(id string) (string, error) {
	c := p.watcher.Container(id)
	if c == nil {
		return "", errors.New("container not found : " + id)
	}
	return c.Config.Labels["com.docker.compose.project"], nil
}

func (p *Projects) Project(b *traefik.Backend) (string, error) {
	for server := range b.Servers {
		fmt.Println(server)
		l := strings.LastIndexByte(server, '-')
		if l < 0 { // It should not append
			continue
		}
		n := server[:l]
		cs, err := p.watcher.Find(func(c *types.ContainerJSON) (bool, error) {
			return NormalizeName(c.Name) == n, nil
		})
		if err != nil {
			return "", err
		}
		if len(cs) > 1 {
			return "", errors.New("Oups, not unique container : " + n)
		}
		// All containers of a backend must be in the same project
		return cs[0].Config.Labels["com.docker.compose.project"], nil
	}
	return "", nil
}
