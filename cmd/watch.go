package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/docker/docker/client"
	"github.com/factorysh/docker-visitor/visitor"
	"github.com/factorysh/traefik-sidecar/events"
	"github.com/factorysh/traefik-sidecar/projects"
	"github.com/factorysh/traefik-sidecar/story"
	"github.com/factorysh/traefik-sidecar/version"
	"github.com/factorysh/traefik-sidecar/web"
	"github.com/onrik/logrus/sentry"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	traefikHost     string
	traefikPassword string
	eventsHost      string
	storyPath       string
)

func init() {
	rootCmd.AddCommand(watchCmd)
	watchCmd.PersistentFlags().StringVarP(&traefikHost, "traefik", "t", "http://localhost:8080", "Træfik admin url")
	watchCmd.PersistentFlags().StringVarP(&traefikPassword, "password", "p", "", "Træfik admin password")
	watchCmd.PersistentFlags().StringVarP(&eventsHost, "events", "e", "localhost:3000", "Events SSE endpoint")
	watchCmd.PersistentFlags().StringVarP(&storyPath, "story", "s", "", "Log backend story to a log path")
}

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "watch for events",
	Long:  ``,
	RunE:  watch,
}

func watch(cmd *cobra.Command, args []string) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	stops := []func(){}

	// logrus hook for sentry, if DSN is provided
	dsn := os.Getenv("SENTRY_DSN")
	if dsn != "" {
		sentryHook, err := sentry.NewHook(sentry.Options{
			Dsn: dsn,
		}, log.PanicLevel, log.FatalLevel, log.ErrorLevel)
		if err != nil {
			return err
		}
		sentryHook.AddTag("version", version.Version())
		sentryHook.AddTag("program", "sidecar")
		log.AddHook(sentryHook)
	}

	docker, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	w := visitor.New(docker)
	ctx, cancel := context.WithCancel(context.Background())
	stops = append(stops, cancel)
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

	if storyPath != "" {
		s := story.New(c.Events, storyPath)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		stops = append(stops, cancel)
		go func() {
			err := s.Listen(ctx)
			if err != nil {
				log.WithError(err).Error()
			}
		}()
	}

	mux := http.NewServeMux()
	ctx2 := context.Background()
	mux.Handle("/events", web.New(ctx2, c.Events))
	go c.WatchBackends()
	log.Info("watch traefik's backends")
	log.Info("Listening HTTP")

	go func() {
		for {
			s := <-signals
			log.WithField("signal", s).Info()
			for _, stop := range stops {
				stop()
			}
			time.Sleep(50 * time.Millisecond)
			os.Exit(1)
		}
	}()
	err = http.ListenAndServe(eventsHost, mux)
	if err != nil {
		log.WithError(err).Error()
	}
	return nil
}
