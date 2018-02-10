package spinnaker

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"

	"github.com/namely/spinnaker-datadog-bridge/spinnaker/types"
)

// Handler defines an interface to allow you to implement your own handlers
// for spinnaker webhooks that have been sent
//go:generate mockgen -package=mocks -destination=./mocks/handler.go github.com/namely/spinnaker-datadog-bridge/spinnaker Handler
type Handler interface {
	Handle(incoming *types.IncomingWebhook) error
}

// HandlerMap contains all of the handlers and the type of detail they are used for
type HandlerMap map[string]Handler

// Dispatcher contains all of the registered handlers for incoming webhooks
// from Spinnaker based on their detail type. For example:
// "orca:stage:complete"
type Dispatcher struct {
	handlers HandlerMap
}

// NewDispatcher initializes a new dispatcher instance
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(HandlerMap),
	}
}

// Handlers returns the current handlers associated with this dispatcher
func (d *Dispatcher) Handlers() HandlerMap {
	return d.handlers
}

// AddHandler adds a handler for the given hook type (orca:stage:complete for example)
func (d *Dispatcher) AddHandler(hookType string, h Handler) {
	d.handlers[hookType] = h
}

// HandleIncomingRequest reads a given http request object and dispatches the
// appropriate handler for it (if any exists). Returns true and nil if the handler
// existed and completed without an error. Return false and nil if the handler did not exist at all.
// or Returns true and an error if the handler existed but failed to complete
func (d *Dispatcher) HandleIncomingRequest(req *http.Request) (exists bool, err error) {
	var incoming types.IncomingWebhook

	if err := json.NewDecoder(req.Body).Decode(&incoming); err != nil {
		return false, errors.Wrap(err, "could not decode incoming webhook")
	}

	handler, exists := d.Handlers()[incoming.Details.Type]
	if !exists {
		return false, nil
	}

	if err := handler.Handle(&incoming); err != nil {
		return false, err
	}

	return false, nil
}
