package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/urfave/cli/v3"
	"go.wasmcloud.dev/wasmbus"
	"go.wasmcloud.dev/wasmbus/events"
)

func eventCommand() *cli.Command {
	return &cli.Command{
		Name:  "events",
		Usage: "listen for lattice events",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "pattern",
				Value: wasmbus.PatternAll,
				Usage: "The event pattern to subscribe to",
			},
			&cli.IntFlag{
				Name:  "backlog",
				Value: wasmbus.NoBackLog,
				Usage: "Bus backlog size. Default is no backlog",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			bus, err := busFromCommand(cmd)
			if err != nil {
				return err
			}

			lattice := cmd.String("lattice")
			pattern := cmd.String("pattern")
			backlog := cmd.Int("backlog")

			callback := func(_ context.Context, ev events.Event) {
				jsonEv, err := json.MarshalIndent(ev.BusEvent, "", "  ")
				if err != nil {
					log.Printf("Error marshalling event: %s", err)
					return
				}

				fmt.Println(titleStyle.Render("⁜", ev.CloudEvent.Type()))
				renderKv("Time", ev.CloudEvent.Time().Format(time.RFC3339))
				renderKv("Source", ev.CloudEvent.Source())
				renderKv("Id", ev.CloudEvent.ID())
				fmt.Println(string(jsonEv))
			}
			subscription, err := events.Subscribe(bus, lattice, pattern, int(backlog), events.DiscardErrorsHandler(callback))
			if err != nil {
				return err
			}
			defer subscription.Drain()

			log.Printf("Listening for events on lattice '%s' with pattern '%s'", lattice, pattern)
			<-ctx.Done()
			log.Printf("Shutting down event listener")

			return nil
		},
	}
}
