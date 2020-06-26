package cmd

import (
	"context"
	"net/http"

	"github.com/docker/docker/client"
	"github.com/factorysh/docker-visitor/visitor"
	"github.com/factorysh/traefik-sidecar/events"
	"github.com/factorysh/traefik-sidecar/projects"
	"github.com/factorysh/traefik-sidecar/web"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	traefikHost     string
	traefikPassword string
	eventsHost      string
)

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.PersistentFlags().StringVarP(&traefikHost, "traefik", "t", "http://localhost:8080", "Træfik admin url")
	watchCmd.PersistentFlags().StringVarP(&traefikPassword, "password", "p", "", "Træfik admin password")
	watchCmd.PersistentFlags().StringVarP(&eventsHost, "events", "e", "localhost:3000", "Events SSE endpoint")
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "watch for events",
	Long:  ``,
	RunE:  watch,
}

func watch(cmd *cobra.Command, args []string) error {
	docker, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	w := visitor.New(docker)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := w.Start(ctx)
		if err != nil {
			log.WithError(err).Error()
		}
	}()
	log.Info("Listen docker events")
	p := projects.New(w)
	c, err := events.New(traefikHost, "admin", traefikPassword, p)
	if err != nil {
		return err
	}
	mux := http.NewServeMux()
	ctx2 := context.Background()
	mux.Handle("/events", web.New(ctx2, c.Events))
	go c.WatchBackends()
	log.Info("watch traefik's backends")
	log.Info("Listening HTTP")
	err = http.ListenAndServe(eventsHost, mux)
	if err != nil {
		log.WithError(err).Error()
	}
	return nil
}
