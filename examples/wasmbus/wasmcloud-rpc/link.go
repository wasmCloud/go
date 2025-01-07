package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/urfave/cli/v3"
	"go.wasmcloud.dev/wasmbus/control"
)

func linkCommand() *cli.Command {
	var targetName string
	var nameArg = &cli.StringArg{
		Name:        "name",
		Destination: &targetName,
		Max:         1,
	}
	return &cli.Command{
		Name:  "link",
		Usage: "Interact with Lattice Links",
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:    "get",
				Aliases: []string{"list"},
				Usage:   "Get all links",
				Action:  getLinkCommand,
			},
			{
				Name:   "delete",
				Usage:  "Delete a link",
				Action: deleteLinkCommand,
				Arguments: []cli.Argument{
					nameArg,
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "name",
						Usage: "Link Name",
						Value: "default",
					},
					&cli.StringFlag{
						Name:     "source",
						Usage:    "Link Source",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "namespace",
						Usage:    "WIT Namespace",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "package",
						Usage:    "WIT Package",
						Required: true,
					},
				},
			},
			{
				Name:   "put",
				Usage:  "Put a new link",
				Action: putLinkCommand,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "namespace",
						Usage:    "WIT Namespace",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "package",
						Usage:    "WIT Package",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:     "interface",
						Usage:    "WIT Interfaces",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "name",
						Usage: "Link Name",
						Value: "default",
					},
					&cli.StringFlag{
						Name:     "source",
						Usage:    "Link Source",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "target",
						Usage:    "Link Target",
						Required: true,
					},
					&cli.StringSliceFlag{
						Name:  "source-config",
						Usage: "Named Configurations for Source",
					},
					&cli.StringSliceFlag{
						Name:  "target-config",
						Usage: "Named Configurations for Target",
					},
				},
				Arguments: []cli.Argument{
					nameArg,
				},
			},
		},
	}
}

func deleteLinkCommand(ctx context.Context, cmd *cli.Command) error {
	name := cmd.String("name")

	client, err := controlClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.LinkDelete(ctx, &control.LinkDeleteRequest{
		Name:         cmd.String("name"),
		SourceId:     cmd.String("source"),
		WitNamespace: cmd.String("namespace"),
		WitPackage:   cmd.String("package"),
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("received error response: %s", resp.Message)
	}

	fmt.Println(titleStyle.Render("⁜", name), "deleted")

	return nil
}

func getLinkCommand(ctx context.Context, cmd *cli.Command) error {
	client, err := controlClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.LinkGet(ctx, &control.LinkGetRequest{})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("received error response: %s", resp.Message)
	}
	t := newTable()
	t.Headers("Name", "Source", "Target", "Namespace", "Package", "Interfaces")

	for _, link := range resp.Response {
		t.Row(link.Name, link.SourceId, link.Target, link.WitNamespace, link.WitPackage, strings.Join(link.WitInterfaces, ","))
	}
	fmt.Println(t)

	return nil
}

func putLinkCommand(ctx context.Context, cmd *cli.Command) error {
	name := cmd.String("name")
	client, err := controlClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.LinkPut(ctx, &control.LinkPutRequest{
		Name:          name,
		SourceId:      cmd.String("source"),
		Target:        cmd.String("target"),
		WitNamespace:  cmd.String("namespace"),
		WitPackage:    cmd.String("package"),
		WitInterfaces: cmd.StringSlice("interface"),
		SourceConfig:  cmd.StringSlice("source-config"),
		TargetConfig:  cmd.StringSlice("target-config"),
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("received error response: %s", resp.Message)
	}

	fmt.Println(titleStyle.Render("⁜", name), "stored")

	return nil
}
