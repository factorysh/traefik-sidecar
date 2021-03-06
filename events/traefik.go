package events

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/crc64"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/factorysh/pubsub/event"
	log "github.com/sirupsen/logrus"

	jsonpatch "github.com/evanphx/json-patch"
	_projects "github.com/factorysh/traefik-sidecar/projects"
	"github.com/factorysh/traefik-sidecar/traefik"
	"github.com/yazgazan/jaydiff/diff"
)

// Client for traefik API
type Client struct {
	req          *http.Request
	address      string
	client       *http.Client
	Events       *event.Events
	currentState []byte
	lock         sync.WaitGroup
	projects     *_projects.Projects
	errorWait    time.Duration
	pollWait     time.Duration
}

// New Client
func New(address, username, password string, projects *_projects.Projects) (*Client, error) {
	req, err := http.NewRequest("GET", address, nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	c := &Client{
		address:   address,
		req:       req,
		client:    &http.Client{},
		Events:    event.NewEvents(),
		projects:  projects,
		errorWait: 5 * time.Second,
		pollWait:  time.Second,
	}
	c.lock.Add(1)
	c.Events.SetPrems(func(ctx context.Context) *event.Event {
		c.lock.Wait()
		return &event.Event{
			Id:    "0",
			Data:  string(c.currentState),
			Event: "initial",
		}
	})
	return c, nil
}

// Wait for getting initial traefik state
func (c *Client) Wait() {
	c.lock.Wait()
}

// WatchBackends watch traefik's backends
// It's polling : forever loop + wait
// Events are diff between current and last state.
func (c *Client) WatchBackends() error {
	table := crc64.MakeTable(42)
	c.req.URL.Path = "/api/providers/docker/backends"
	var ckOld uint64
	for {
		resp, err := c.client.Do(c.req)
		if err != nil {
			log.Error(err)
			time.Sleep(c.errorWait)
			continue
		}
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusBadRequest {
			// Blocking error
			return fmt.Errorf("Traefik deadly error : %s", resp.Status)
		}
		if resp.StatusCode != http.StatusOK {
			// Not blocking error
			log.WithField("status", resp.StatusCode).Error()
			time.Sleep(c.errorWait)
			continue
		}
		bodyText, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			continue
		}
		ck := crc64.Checksum(bodyText, table)
		if len(c.currentState) == 0 {
			var current traefik.Backends
			err = json.Unmarshal(bodyText, &current)
			if err != nil {
				log.Error(err)
			}
			l := log.WithField("ck", ck)
			for name, backend := range current {
				p, err := c.projects.Project(backend)
				if err != nil {
					log.Error(err)
				} else {
					l = l.WithField(name, p)
					log.Info("Server ", name, "project", p)
				}
			}
			l.Info("Initial state")
			c.currentState = bodyText
			c.lock.Done()
		} else {
			if ck != ckOld {
				var a1 map[string]interface{}
				var a2 map[string]interface{}

				json.Unmarshal(c.currentState, &a1)
				json.Unmarshal(bodyText, &a2)
				d, _ := diff.Diff(a1, a2)
				fmt.Println(d.StringIndent("", "", diff.Output{
					Indent:    "  ",
					ShowTypes: true,
					Colorized: true,
				}))

				patch, err := jsonpatch.CreateMergePatch(c.currentState, bodyText)
				if err != nil {
					log.Error(err)
				} else {
					c.Events.Append(&event.Event{
						Data:  string(patch),
						Id:    c.Events.NextEventId(),
						Event: "traefik.patch",
					})
				}
				c.currentState = bodyText
			}
		}
		ckOld = ck
		time.Sleep(c.pollWait)
	}
}
