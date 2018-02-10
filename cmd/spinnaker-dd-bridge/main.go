package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
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

	http.HandleFunc("/webhook/", func(w http.ResponseWriter, req *http.Request) {
		var webhook Webhook
		if err := json.NewDecoder(req.Body).Decode(&webhook); err != nil {
			logrus.WithError(err).Error()
		}

		if webhook.Details.Type == "orca:pipeline:starting" {
			e := &datadog.Event{}
			e.SetTitle("Pipeline Started")
			e.SetAggregation(webhook.Content.Execution.ID)
			e.Tags = []string{
				fmt.Sprintf("app:%s", webhook.Details.Application),
				"pipeline:starting",
			}
			e.SetText("# Pipeline Update!")

			_, err := ddClient.PostEvent(e)
			if err != nil {
				logrus.WithError(err).Error("could not post pipeline update to datadog")
			}
		}

		w.WriteHeader(http.StatusOK)
	})

	return http.ListenAndServe(":1991", http.DefaultServeMux)
}
