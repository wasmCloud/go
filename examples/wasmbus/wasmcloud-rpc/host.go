package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/urfave/cli/v3"
	"go.wasmcloud.dev/x/wasmbus/control"
)

func hostCommand() *cli.Command {
	var targetName string
	nameArg := &cli.StringArg{
		Name:        "host",
		Destination: &targetName,
		Max:         1,
	}
	return &cli.Command{
		Name:  "host",
		Usage: "Host management commands",
		Commands: []*cli.Command{
			{
				Name:   "stop",
				Usage:  "Stop a host",
				Action: wrapNamedAction(stopHostCommand, &targetName),
				Arguments: []cli.Argument{
					nameArg,
				},
			},
			{
				Name:   "put-label",
				Usage:  "Put a label on a host",
				Action: wrapNamedAction(putLabelCommand, &targetName),
				Arguments: []cli.Argument{
					nameArg,
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "key",
						Usage:    "Label key",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "value",
						Usage:    "Label value",
						Required: true,
					},
				},
			},
			{
				Name:   "delete-label",
				Usage:  "Delete a label from a host",
				Action: wrapNamedAction(deleteLabelCommand, &targetName),
				Arguments: []cli.Argument{
					nameArg,
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "key",
						Usage:    "Label key",
						Required: true,
					},
				},
			},
			{
				Name:   "inventory",
				Usage:  "Get a host inventory",
				Action: wrapNamedAction(inventoryCommand, &targetName),
				Arguments: []cli.Argument{
					nameArg,
				},
			},
			{
				Name:   "ping",
				Usage:  "Ping all hosts",
				Action: pingCommand,
			},
		},
	}
}

func stopHostCommand(ctx context.Context, cmd *cli.Command, host string) error {
	client, err := controlClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.HostStop(ctx, &control.HostStopRequest{
		HostId: host,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("received error response: %s", resp.Message)
	}

	fmt.Println(titleStyle.Render("⁜", host), "stopped")
	return nil
}

func putLabelCommand(ctx context.Context, cmd *cli.Command, host string) error {
	client, err := controlClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.HostLabelPut(ctx, &control.HostLabelPutRequest{
		HostId: host,
		Key:    cmd.String("key"),
		Value:  cmd.String("value"),
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("received error response: %s", resp.Message)
	}

	fmt.Println(titleStyle.Render("⁜", host), "label updated")
	return nil
}

func deleteLabelCommand(ctx context.Context, cmd *cli.Command, host string) error {
	client, err := controlClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.HostLabelDelete(ctx, &control.HostLabelDeleteRequest{
		HostId: host,
		Key:    cmd.String("key"),
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("received error response: %s", resp.Message)
	}

	fmt.Println(titleStyle.Render("⁜", host), "label deleted")
	return nil
}

func inventoryCommand(ctx context.Context, cmd *cli.Command, host string) error {
	client, err := controlClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.HostInventory(ctx, &control.HostInventoryRequest{
		HostId: host,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("received error response: %s", resp.Message)
	}

	fmt.Println(titleStyle.Render("⁜", resp.Response.HostId), resp.Response.FriendlyName)

	renderKv("Version", resp.Response.Version)
	renderKv("Uptime", resp.Response.UptimeHuman)
	renderKv("Labels", "")
	for _, l := range fancyLabels(resp.Response.Labels) {
		fmt.Println("  ", l)
	}
	fmt.Println()
	if len(resp.Response.Components) > 0 {
		renderKv("Components", "")
		t := newTable()
		t.Headers("Name", "Instances", "Revision", "Image")
		for _, c := range resp.Response.Components {
			t.Row(c.Name, strconv.FormatInt(int64(c.MaxInstances), 10), strconv.FormatInt(int64(c.Revision), 10), c.ImageRef)
		}
		fmt.Println(t)
	}

	if len(resp.Response.Providers) > 0 {
		renderKv("Providers", "")
		t := newTable()
		t.Headers("Name", "Revision", "Image")
		for _, p := range resp.Response.Providers {
			t.Row(p.Name, strconv.FormatInt(int64(p.Revision), 10), p.ImageRef)
		}
		fmt.Println(t)
	}
	return nil
}

func pingCommand(ctx context.Context, cmd *cli.Command) error {
	client, err := controlClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.HostPing(ctx, &control.HostPingRequest{
		// NOTE(lxf): We might miss hosts that take too long to reply
		// but we can't wait forever for a response.
		// It's an API flaw.
		Wait: 1 * time.Second,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("received error response: %s", resp.Message)
	}

	for _, r := range resp.Response {
		fmt.Println(titleStyle.Render("⁜", r.Id), r.FriendlyName)
		renderKv("Version", r.Version)
		renderKv("Uptime", r.UptimeHuman)
		renderKv("Labels", "")
		for _, l := range fancyLabels(r.Labels) {
			fmt.Println("  ", l)
		}
		fmt.Println()
	}

	return nil
}

func fancyLabels(labels map[string]string) []string {
	var out []string
	for k, v := range labels {
		out = append(out, fmt.Sprintf("%s=%s", keyStyle.Render(k), valueStyle.Render(v)))
	}
	return out
}
