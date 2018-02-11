package spinnaker

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/bobbytables/spinnaker-datadog-bridge/spinnaker/types"
)

// Handler defines an interface to allow you to implement your own handlers
// for spinnaker webhooks that have been sent
//go:generate mockgen -package=mocks -destination=./mocks/handler.go github.com/bobbytables/spinnaker-datadog-bridge/spinnaker Handler
type Handler interface {
	Handle(incoming *types.IncomingWebhook) error
	Name() string
}

// HandlerMap contains all of the handlers and the type of detail they are used for
type HandlerMap map[string][]Handler

// Dispatcher contains all of the registered handlers for incoming webhooks
// from Spinnaker based on their detail type. For example:
// "orca:stage:complete"
type Dispatcher struct {
	handlers HandlerMap
}

// DispatchResult is returned from the webhook handler onto a channel
// to allow piping multiple handlers per hook but still get insight into
// the result of each one so you can error or log
type DispatchResult struct {
	HookType    string
	HandlerName string
	Err         error
	Duration    time.Duration
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
	if _, ok := d.handlers[hookType]; !ok {
		d.handlers[hookType] = make([]Handler, 0)
	}

	d.handlers[hookType] = append(d.handlers[hookType], h)
}

// HandleIncomingRequest reads a given http request object and dispatches the
// appropriate handlers for it (if any exists). If it fails to decode the
// incoming request body it will return an error. Otherwise, a channel is returned
// that results are sent to as the given handlers complete or fail.
func (d *Dispatcher) HandleIncomingRequest(req *http.Request) (<-chan DispatchResult, error) {
	incoming := new(types.IncomingWebhook)

	if err := json.NewDecoder(req.Body).Decode(incoming); err != nil {
		return nil, errors.Wrap(err, "could not decode incoming webhook")
	}

	handlers := d.Handlers()[incoming.Details.Type]

	var wg sync.WaitGroup
	wg.Add(len(handlers))

	results := make(chan DispatchResult)
	for _, handler := range handlers {
		go func(handler Handler) {
			start := time.Now()
			err := handler.Handle(incoming)
			took := time.Since(start)
			results <- DispatchResult{
				Err:         err,
				HandlerName: handler.Name(),
				HookType:    incoming.Details.Type,
				Duration:    took,
			}

			wg.Done()
		}(handler)
	}

	// Once we've processed all of our handlers we're going to close the channel
	// so receivers can act accordingly
	go func() {
		wg.Wait()
		close(results)
	}()

	return results, nil
}
