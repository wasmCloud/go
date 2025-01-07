package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/urfave/cli/v3"
	"go.wasmcloud.dev/wasmbus"
	"go.wasmcloud.dev/wasmbus/control"
	"go.wasmcloud.dev/wasmbus/wadm"
)

var (
	titleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("35"))
	keyStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	oddStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("35"))
	evenStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("83"))
)

func renderKv(k, v string) {
	fmt.Println(keyStyle.Render(k), valueStyle.Render(v))
}

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "lattice",
				Value: "default",
				Usage: "The lattice ID to use",
			},
			&cli.StringFlag{
				Name:  "nats-url",
				Value: wasmbus.NatsDefaultURL,
				Usage: "The NATS server URL",
			},
		},
		Commands: []*cli.Command{
			eventCommand(),
			wadmCommand(),
			configCommand(),
			linkCommand(),
			hostCommand(),
		},
	}

	signalCh, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := cmd.Run(signalCh, os.Args); err != nil {
		log.Fatal(err)
	}
}

func busFromCommand(cmd *cli.Command) (wasmbus.Bus, error) {
	nc, err := wasmbus.NatsConnect(cmd.String("nats-url"))
	if err != nil {
		return nil, err
	}

	return wasmbus.NewNatsBus(nc), nil
}

func wadmClientFromCommand(cmd *cli.Command) (*wadm.Client, error) {
	bus, err := busFromCommand(cmd)
	if err != nil {
		return nil, err
	}
	lattice := cmd.String("lattice")
	return wadm.NewClient(bus, lattice), nil
}

func controlClientFromCommand(cmd *cli.Command) (*control.Client, error) {
	bus, err := busFromCommand(cmd)
	if err != nil {
		return nil, err
	}
	lattice := cmd.String("lattice")
	return control.NewClient(bus, lattice), nil
}

func newTable() *table.Table {
	return table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			switch {
			case row%2 == 0:
				return evenStyle
			default:
				return oddStyle
			}
		})
}

type namedActionFunc func(ctx context.Context, cmd *cli.Command, name string) error

func wrapNamedAction(f namedActionFunc, name *string) cli.ActionFunc {
	return func(ctx context.Context, cmd *cli.Command) error {
		if name == nil || *name == "" {
			return fmt.Errorf("name is required")
		}
		return f(ctx, cmd, *name)
	}
}
