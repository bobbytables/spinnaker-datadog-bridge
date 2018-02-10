package spinnakerdatadog

import (
	"html/template"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
	"github.com/namely/spinnaker-datadog-bridge/spinnaker"
	"github.com/pkg/errors"
	datadog "gopkg.in/zorkian/go-datadog-api.v2"
)

// Spout is the main handler for all of the Spinnaker events. It attaches
// handlers to a Spinnaker dispatcher to fan out events correctly
type Spout struct {
	client         *datadog.Client
	eventTemplates map[string]*EventTemplate
}

// EventTemplate is the representation in the template file
// before parsing it
type EventTemplate struct {
	Title string `json:"title,omitempty"`
	Text  string `json:"text,omitempty"`

	compiledTitle *template.Template
	compiledText  *template.Template
	isCompiled    bool
}

func (et *EventTemplate) Compile() error {
	if et.isCompiled {
		return nil
	}

	var err error
	et.compiledTitle, err = template.New("eventTitle").Parse(et.Title)
	if err != nil {
		return errors.Wrap(err, "could not compile eventTitle")
	}

	et.compiledText, err = template.New("eventText").Parse(et.Text)
	if err != nil {
		return errors.Wrap(err, "could not compile eventText")
	}

	return err
}

// NewSpout initializes a new spout for spitting out datadog events from
// Spinnaker event webhooks
func NewSpout(c *datadog.Client, templateFile string) (*Spout, error) {
	spout := &Spout{client: c}

	if templateFile == "" {
		return spout, nil
	}

	f, err := os.Open(templateFile)
	if err != nil {
		return nil, errors.Wrap(err, "could not open template file")
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Wrap(err, "could not read template file")
	}

	et := make(map[string]*EventTemplate)
	if err := yaml.Unmarshal(b, &et); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal template file")
	}

	spout.eventTemplates = et
	return spout, nil
}

// TotalTemplates returns how many templates are currently registered
// for events
func (s *Spout) TotalTemplates() int {
	return len(s.eventTemplates)
}

// Handlers returns the handlers that get attached to a dispatcher when AttachToDispatcher is called
func (s *Spout) Handlers() map[string][]spinnaker.Handler {
	hs := make(map[string][]spinnaker.Handler)

	for hookType, eventTemplate := range s.eventTemplates {
		hs[hookType] = []spinnaker.Handler{
			&DatadogEventHandler{spout: s, template: eventTemplate},
		}
	}

	return hs
}

// AttachToDispatcher registers all of the handlers for this spout to a spinnaker
// dispatcher.
func (s *Spout) AttachToDispatcher(d *spinnaker.Dispatcher) {
	for hookType, handlers := range s.Handlers() {
		for _, handler := range handlers {
			d.AddHandler(hookType, handler)
		}
	}
}
