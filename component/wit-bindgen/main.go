package main

import (
	"context"
	"fmt"
	"os"

	"github.com/urfave/cli/v3"

	"go.bytecodealliance.org/cmd/wit-bindgen-go/cmd/generate"
)

// wrapper around bytecodealliance wit-bindgen-go
// Uses the component sdk 'cm' package if one is not provided
// prevents cm package version mismatches at build time

func main() {
	cmd := generateCommand()
	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func generateCommand() *cli.Command {
	gen := generate.Command
	gen.Name = "wasmcloud-wit-bindgen"
	gen.Before = func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		if cmd.String("cm") == "" {
			if err := gen.Set("cm", "go.wasmcloud.dev/component/cm"); err != nil {
				return ctx, err
			}
		}

		return ctx, nil
	}

	return gen
}
