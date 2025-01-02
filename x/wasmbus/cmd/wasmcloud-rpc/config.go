package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v3"
	"go.wasmcloud.dev/wasmbus/control"
)

func configCommand() *cli.Command {
	var targetName string
	var nameArg = &cli.StringArg{
		Name:        "name",
		Destination: &targetName,
		Max:         1,
	}
	return &cli.Command{
		Name:  "config",
		Usage: "Interact with Lattice Config",
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:   "get",
				Usage:  "Get a config",
				Action: wrapNamedAction(getConfigCommand, &targetName),
				Arguments: []cli.Argument{
					nameArg,
				},
			},
			{
				Name:   "delete",
				Usage:  "Delete a config",
				Action: wrapNamedAction(deleteConfigCommand, &targetName),
				Arguments: []cli.Argument{
					nameArg,
				},
			},
			{
				Name:   "put",
				Usage:  "Put a config",
				Action: putConfigCommand,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Name of the config to store",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "File to read config from, in dotenv format (one KEY=VALUE per line)",
						Required: true,
					},
				},
				Arguments: []cli.Argument{
					nameArg,
				},
			},
		},
	}
}

func deleteConfigCommand(ctx context.Context, cmd *cli.Command, name string) error {
	client, err := controlClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.ConfigDelete(ctx, &control.ConfigDeleteRequest{
		Name: name,
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

func getConfigCommand(ctx context.Context, cmd *cli.Command, name string) error {
	client, err := controlClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.ConfigGet(ctx, &control.ConfigGetRequest{
		Name: name,
	})
	if err != nil {
		return err
	}
	if !resp.Success {
		return fmt.Errorf("received error response: %s", resp.Message)
	}

	// NOTE(lxf): This feels like a bug in hosts' crates/host/src/wasmbus/mod.rs
	if resp.Message == "Configuration not found" {
		return fmt.Errorf("configuration not found")
	}

	config, err := godotenv.Marshal(resp.Response)
	if err != nil {
		return err
	}
	fmt.Println(config)

	return nil
}

func putConfigCommand(ctx context.Context, cmd *cli.Command) error {
	name := cmd.String("name")
	client, err := controlClientFromCommand(cmd)
	if err != nil {
		return err
	}

	f, err := os.ReadFile(cmd.String("file"))
	if err != nil {
		return err
	}

	var values map[string]string
	if values, err = godotenv.UnmarshalBytes(f); err != nil {
		return err
	}

	resp, err := client.ConfigPut(ctx, &control.ConfigPutRequest{
		Name:   name,
		Values: values,
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
