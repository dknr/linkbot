package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"maunium.net/go/mautrix/event"

	"ragecore"
)

func main() {
	configPath := flag.String("config", "config.json", "path to the configuration file")
	flag.Parse()

	cfg, err := ragecore.LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := zerolog.New(zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.TimeFormat = time.StampMilli
	})).With().Timestamp().Logger()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	bot, err := ragecore.New(ctx, cfg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start bot")
	}

	// Preview the first allowed link in every incoming message. OnMessage skips
	// our own sends so we do not re-preview what we replied with.
	bot.OnMessage(func(ctx context.Context, evt *event.Event) {
		body := evt.Content.AsMessage().Body
		if body == "" {
			return
		}
		link, ok := firstAllowedLink(body)
		if !ok {
			return
		}
		meta, err := fetchOG(ctx, link)
		if err != nil {
			log.Error().Err(err).
				Str("url", link).
				Stringer("room_id", evt.RoomID).
				Stringer("event_id", evt.ID).
				Msg("Failed to fetch Open Graph")
			return
		}
		log.Debug().
			Str("url", link).
			Str("title", meta.Title).
			Stringer("room_id", evt.RoomID).
			Stringer("event_id", evt.ID).
			Msg("Sending link preview")
		content := &event.MessageEventContent{
			MsgType: event.MsgNotice,
			Body:    formatPreview(meta),
		}
		if _, err := bot.Client().SendMessageEvent(ctx, evt.RoomID, event.EventMessage, content); err != nil {
			log.Error().Err(err).Stringer("room_id", evt.RoomID).Msg("Failed to send link preview")
		}
	})

	if err := bot.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("Bot failed")
	}
}
