package spinnakerdatadog

import (
	"bytes"
	"fmt"

	"github.com/namely/spinnaker-datadog-bridge/spinnaker"
	"github.com/namely/spinnaker-datadog-bridge/spinnaker/types"
	"github.com/pkg/errors"
	datadog "gopkg.in/zorkian/go-datadog-api.v2"
)

// DatadogEventHandler handles piping all of the registered events (via templates)
// to datadog when the dispatcher receives them. It handles compiling the original template
// and then sending it to the DataDog events API
type DatadogEventHandler struct {
	spout    *Spout
	template *EventTemplate
}

var _ spinnaker.Handler = (*DatadogEventHandler)(nil)

func NewDatadogEventHandler(s *Spout, template *EventTemplate) *DatadogEventHandler {
	return &DatadogEventHandler{
		spout:    s,
		template: template,
	}
}

// Handle implements spinnaker.Handler. It sends datadog events for the given
// webhook event type. It compiles the given template from the webhook and sends it
func (deh *DatadogEventHandler) Handle(incoming *types.IncomingWebhook) error {
	if err := deh.template.Compile(); err != nil {
		return errors.Wrap(err, "could not compile template")
	}

	titleBuf, textBuf := new(bytes.Buffer), new(bytes.Buffer)
	if err := deh.template.compiledTitle.Execute(titleBuf, incoming); err != nil {
		return errors.Wrap(err, "could not compile title from webhook")
	}

	if err := deh.template.compiledText.Execute(textBuf, incoming); err != nil {
		return errors.Wrap(err, "could not compile text from webhook")
	}

	event := &datadog.Event{}
	event.SetTitle(titleBuf.String())
	event.SetText(textBuf.String())
	event.SetAggregation(incoming.Content.ExecutionID)
	event.Tags = []string{
		fmt.Sprintf("app:%s", incoming.Details.Application),
		incoming.Details.Type,
	}

	if _, err := deh.spout.client.PostEvent(event); err != nil {
		return errors.Wrap(err, "could not post to datadog API")
	}

	return nil
}
