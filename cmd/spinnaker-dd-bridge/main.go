package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	datadog "gopkg.in/zorkian/go-datadog-api.v2"

	"github.com/bobbytables/spinnaker-datadog-bridge/server"
	"github.com/bobbytables/spinnaker-datadog-bridge/spinnaker"
	"github.com/bobbytables/spinnaker-datadog-bridge/spinnakerdatadog"
)

func main() {
	app := cli.NewApp()
	app.Name = "spinnaker-dd-bridge"
	app.Action = serverAction
	app.Authors = []cli.Author{
		{
			Name:  "Robert Ross",
			Email: "robert.ross@namely.com",
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "datadog-api-key",
			Usage:  "your datadog api key (Found at https://app.datadoghq.com/account/settings#api)",
			EnvVar: "DATADOG_API_KEY",
		},
		cli.StringFlag{
			Name:   "datadog-app-key",
			Usage:  "your datadog app key (Found at https://app.datadoghq.com/account/settings#api)",
			EnvVar: "DATADOG_APP_KEY",
		},
		cli.StringFlag{
			Name:   "event-templates",
			Usage:  "The file where your event templates are located for Spinnaker events",
			EnvVar: "EVENT_TEMPLATES",
		},
		cli.StringFlag{
			Name:   "addr",
			Usage:  "The address the server listens on",
			EnvVar: "ADDR",
			Value:  ":3000",
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "Turn on DEBUG level logging",
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err.Error())
		os.Exit(1)
	}
}

func serverAction(c *cli.Context) error {
	ddClient := datadog.NewClient(c.String("datadog-api-key"), c.String("datadog-app-key"))
	dispatcher := spinnaker.NewDispatcher()
	spout, err := spinnakerdatadog.NewSpout(ddClient, c.String("event-templates"))
	if err != nil {
		return err
	}

	if c.Bool("debug") {
		logrus.StandardLogger().SetLevel(logrus.DebugLevel)
	}

	spout.AttachToDispatcher(dispatcher)

	return server.New(c.String("addr"), dispatcher).Start()
}
