package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/namely/spinnaker-datadog-bridge/spinnaker"
	"github.com/namely/spinnaker-datadog-bridge/spinnakerdatadog"
	"github.com/urfave/cli"

	datadog "gopkg.in/zorkian/go-datadog-api.v2"
)

type Webhook struct {
	Details Details `json:"details"`
	Content Content `json:"content"`
}

type Details struct {
	Source      string `json:"source"`
	Type        string `json:"type"`
	Created     string `json:"created"`
	Application string `json:"application"`
}

type Content struct {
	Execution Execution `json:"execution"`
}

type Execution struct {
	ID string `json:"id"`
}

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

	spout.AttachToDispatcher(dispatcher)

	http.HandleFunc("/webhook/", func(w http.ResponseWriter, req *http.Request) {
		dispatcher.HandleIncomingRequest(req)
	})

	return http.ListenAndServe(":1991", http.DefaultServeMux)
}
