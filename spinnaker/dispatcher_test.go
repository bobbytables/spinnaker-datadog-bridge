package spinnaker_test

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bobbytables/spinnaker-datadog-bridge/spinnaker"
	"github.com/bobbytables/spinnaker-datadog-bridge/spinnaker/mocks"
	"github.com/bobbytables/spinnaker-datadog-bridge/spinnaker/types"
)

func TestDispatcherAddsHandlers(t *testing.T) {
	d := spinnaker.NewDispatcher()
	h := &mocks.MockHandler{}
	d.AddHandler("test", h)

	assert.Len(t, d.Handlers(), 1)
}

func TestDispatcherHandlesRequests(t *testing.T) {
	tests := []handlerTest{
		{
			scenario:        "Webhook JSON is valid and dispatches the hook",
			requestBodyFile: "valid-webhook.json",
			hookType:        "orca:stage:complete",
			mockFactory: func(ctrl *gomock.Controller, t *testing.T) *mocks.MockHandler {
				m := mocks.NewMockHandler(ctrl)
				m.EXPECT().Handle(gomock.Any()).Do(func(incoming *types.IncomingWebhook) {
					assert.Equal(t, "orca:stage:complete", incoming.Details.Type)
				})

				return m
			},
			assertion: func(d *spinnaker.Dispatcher, req *http.Request, t *testing.T) {
				results, err := d.HandleIncomingRequest(req)
				require.NoError(t, err)

				select {
				case result := <-results:
					require.NoError(t, result.Err)
				case <-time.After(time.Millisecond * 100):
					t.Error("channel never closed")
				}
			},
		},
		{
			scenario:        "The handler fails to handle the incoming webhook",
			requestBodyFile: "valid-webhook.json",
			hookType:        "orca:stage:complete",
			mockFactory: func(ctrl *gomock.Controller, t *testing.T) *mocks.MockHandler {
				m := mocks.NewMockHandler(ctrl)
				m.EXPECT().Handle(gomock.Any()).Do(func(incoming *types.IncomingWebhook) {
					assert.Equal(t, "orca:stage:complete", incoming.Details.Type)
				}).Return(errors.New("well that sucks"))

				return m
			},
			assertion: func(d *spinnaker.Dispatcher, req *http.Request, t *testing.T) {
				results, err := d.HandleIncomingRequest(req)
				require.NoError(t, err)

				select {
				case result := <-results:
					require.Error(t, result.Err)
				case <-time.After(time.Millisecond * 100):
					t.Error("channel never closed")
				}
			},
		},
		{
			scenario:        "Invalid JSON bubbles an error up from the dispatcher",
			requestBodyFile: "bunk-data.json",
			hookType:        "non:existent",
			mockFactory:     defaultMockFactory,
			assertion: func(d *spinnaker.Dispatcher, req *http.Request, t *testing.T) {
				results, err := d.HandleIncomingRequest(req)
				require.Error(t, err)
				require.Nil(t, results)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			handler := test.mockFactory(ctrl, t)
			d := spinnaker.NewDispatcher()
			d.AddHandler(test.hookType, handler)
			test.assertion(d, requestFromFile(test.requestBodyFile), t)
		})
	}
}

type handlerTest struct {
	scenario        string
	mockFactory     func(ctrl *gomock.Controller, t *testing.T) *mocks.MockHandler
	assertion       func(d *spinnaker.Dispatcher, req *http.Request, t *testing.T)
	hookType        string
	requestBodyFile string
}

func requestFromFile(f string) *http.Request {
	wd, _ := os.Getwd()
	file, err := os.Open(filepath.Join(wd, "testdata", f))
	if err != nil {
		log.Fatal("could not open request file: " + err.Error())
	}

	req, err := http.NewRequest("POST", "/bunk", file)
	if err != nil {
		log.Fatal("could not generate request: " + err.Error())
	}

	return req
}

func defaultMockFactory(ctrl *gomock.Controller, t *testing.T) *mocks.MockHandler {
	m := mocks.NewMockHandler(ctrl)
	return m
}
