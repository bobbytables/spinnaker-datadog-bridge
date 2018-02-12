# Spinnaker -> DataDog Bridge

Spinnaker Datadog Bridge is a project to allow you to pipe Spinnaker events
into DataDog easily. You can send pipeline events, stage events, or orca task events to handlers
that can manipulate the data and send it to DataDog as metrics or events easily.

## Running

This is a single command binary that runs an HTTP server that accepts webhooks at `/webhook`.

```
$ spinnaker-dd-bridge \
  --addr=:3000 \
  --datadog-api-key=<api key> \
  --datadog-app-key=<app key> \
  --event-templates=./event-templates.yml
```

This starts a server on port 3000 with the template defined at `event-templates.yml`. This file is how this application determines which events to send and their format.

An example template file for events looks like:

```
orca:stage:complete:
  title: "{{ .Details.Application }} Stage Completed"
  text: |
    Started at: {{ .Content.StartTime }}
    Finished at: {{ .Content.EndTime }}
```

The key `orca:stage:complete` is the mapping between a Spinnaker event and the given template for title and text. The title and text are what are displayed inside of DataDog. You have access to all of the properties defined in the [IncomingWebhook Struct](spinnaker/types/webhooks.go).
