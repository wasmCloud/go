package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/urfave/cli/v3"
	"go.wasmcloud.dev/x/wasmbus"
	"go.wasmcloud.dev/x/wasmbus/wadm"
)

func wadmCommand() *cli.Command {
	var targetName string
	nameArg := &cli.StringArg{
		Name:        "name",
		Destination: &targetName,
		Max:         1,
	}
	return &cli.Command{
		Name:  "wadm",
		Usage: "Interact with wasmcloud admin",
		Flags: []cli.Flag{},
		Commands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "List all models",
				Action: wadmListCommand,
			},
			{
				Name:  "get",
				Usage: "Get a model",
				Arguments: []cli.Argument{
					nameArg,
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "output",
						Aliases: []string{"o"},
						Value:   "yaml",
						Usage:   "Output format",
						Validator: func(s string) error {
							if s != "yaml" && s != "json" {
								return fmt.Errorf("format must be yaml or json")
							}
							return nil
						},
					},
				},
				Action: wrapNamedAction(wadmGetCommand, &targetName),
			},
			{
				Name:  "versions",
				Usage: "Get a model's versions",
				Arguments: []cli.Argument{
					nameArg,
				},
				Action: wrapNamedAction(wadmVersionsCommand, &targetName),
			},
			{
				Name:  "status",
				Usage: "Get a model's status",
				Arguments: []cli.Argument{
					nameArg,
				},
				Action: wrapNamedAction(wadmStatusCommand, &targetName),
			},
			{
				Name:  "delete",
				Usage: "Delete  a model",
				Arguments: []cli.Argument{
					nameArg,
				},
				Action: wrapNamedAction(wadmDeleteCommand, &targetName),
			},
			{
				Name:  "deploy",
				Usage: "Deploy a previously published model",
				Arguments: []cli.Argument{
					nameArg,
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "version",
						Aliases: []string{"v"},
						Usage:   "Version to deploy",
					},
				},
				Action: wrapNamedAction(wadmDeployCommand, &targetName),
			},
			{
				Name:  "undeploy",
				Usage: "Undeploy a model",
				Arguments: []cli.Argument{
					nameArg,
				},
				Action: wrapNamedAction(wadmUndeployCommand, &targetName),
			},
			{
				Name:  "put",
				Usage: "Put a model without deploying it",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "file",
						Aliases: []string{"f"},
						Usage:   "Path to the manifest file",
					},
				},
				Action: wadmPutCommand,
			},
		},
	}
}

func wadmPutCommand(ctx context.Context, cmd *cli.Command) error {
	client, err := wadmClientFromCommand(cmd)
	if err != nil {
		return err
	}

	manifest, err := wadm.LoadManifest(cmd.String("file"))
	if err != nil {
		return err
	}

	resp, err := client.ModelPut(ctx, &wadm.ModelPutRequest{
		Manifest: *manifest,
	})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("error: %s", resp.Message)
	}

	fmt.Println(titleStyle.Render("⁜", resp.Name), resp.CurrentVersion)

	return nil
}

func wadmDeleteCommand(ctx context.Context, cmd *cli.Command, name string) error {
	client, err := wadmClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.ModelDelete(ctx, &wadm.ModelDeleteRequest{
		Name: name,
	})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("error: %s", resp.Message)
	}

	fmt.Println(titleStyle.Render("⁜", name), "deleted")

	return nil
}

func wadmDeployCommand(ctx context.Context, cmd *cli.Command, name string) error {
	client, err := wadmClientFromCommand(cmd)
	if err != nil {
		return err
	}

	version := wadm.LatestVersion
	if cmd.IsSet("version") {
		version = cmd.String("version")
	}

	resp, err := client.ModelDeploy(ctx, &wadm.ModelDeployRequest{
		Name:    name,
		Version: version,
	})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("error: %s", resp.Message)
	}

	fmt.Println(titleStyle.Render("⁜", resp.Name), "deployed version", titleStyle.Render(resp.Version))

	return nil
}

func wadmUndeployCommand(ctx context.Context, cmd *cli.Command, name string) error {
	client, err := wadmClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.ModelUndeploy(ctx, &wadm.ModelUndeployRequest{
		Name: name,
	})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("error: %s", resp.Message)
	}

	fmt.Println(titleStyle.Render("⁜", resp.Name), "undeployed")

	return nil
}

func wadmGetCommand(ctx context.Context, cmd *cli.Command, name string) error {
	client, err := wadmClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.ModelGet(ctx, &wadm.ModelGetRequest{
		Name: name,
	})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("error: %s", resp.Message)
	}

	mimeType := ""
	switch cmd.String("output") {
	case "yaml":
		mimeType = "application/yaml"
	case "json":
		mimeType = "application/json"
	}
	data, err := wasmbus.EncodeMimetype(resp.Manifest, mimeType)
	if err != nil {
		return err
	}
	fmt.Println(string(data))

	return nil
}

func wadmListCommand(ctx context.Context, cmd *cli.Command) error {
	client, err := wadmClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.ModelList(ctx, &wadm.ModelListRequest{})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("error: %s", resp.Message)
	}

	t := newTable().Headers("Name", "Version", "DeployedVersion", "Status", "Description")

	for _, m := range resp.Models {
		fmt.Println(m.Name)
		status := "unknown"
		if m.DetailedStatus != nil {
			status = string(m.DetailedStatus.Info.Type)
		}
		t.Row(m.Name, m.Version, m.DeployedVersion, status, m.Description)
	}
	fmt.Println(t)

	return nil
}

func wadmVersionsCommand(ctx context.Context, cmd *cli.Command, name string) error {
	client, err := wadmClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.ModelVersions(ctx, &wadm.ModelVersionsRequest{
		Name: name,
	})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("error: %s", resp.Message)
	}

	fmt.Println(titleStyle.Render("⁜", name))

	t := newTable().Headers("Version", "Deployed")

	for _, m := range resp.Versions {
		t.Row(m.Version, strconv.FormatBool(m.Deployed))
	}
	fmt.Println(t)

	return nil
}

func wadmStatusCommand(ctx context.Context, cmd *cli.Command, name string) error {
	client, err := wadmClientFromCommand(cmd)
	if err != nil {
		return err
	}

	resp, err := client.ModelStatus(ctx, &wadm.ModelStatusRequest{
		Name: name,
	})
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("error: %s", resp.Message)
	}

	fmt.Println(titleStyle.Render("⁜", name), resp.Status.Status.Type)

	t := newTable().Headers("Scaler", "Status", "Name")
	for _, s := range resp.Status.Scalers {
		t.Row(s.Kind, string(s.Status.Type), s.Name)
	}

	fmt.Println(t)

	return nil
}
