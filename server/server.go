package server

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/bobbytables/spinnaker-datadog-bridge/spinnaker"
)

// Server handles incoming webhook requests and pushes them to a dispatcher
type Server struct {
	Addr       string
	mux        *mux.Router
	dispatcher *spinnaker.Dispatcher
}

// New initializes and returns a server that will listen on the given address
// and dispatch events from incoming webhooks
func New(address string, d *spinnaker.Dispatcher) *Server {
	return &Server{
		Addr:       address,
		dispatcher: d,
	}
}

// Start starts a server to accept Spinnaker webhook events
func (s *Server) Start() error {
	s.prepare()
	logrus.WithField("addr", s.Addr).Info("starting server")

	return http.ListenAndServe(s.Addr, s.mux)
}

func (s *Server) prepare() {
	if s.mux != nil {
		return
	}

	router := mux.NewRouter()
	router.HandleFunc("/webhook", s.handleWebhook)

	s.mux = router
}

func (s *Server) handleWebhook(w http.ResponseWriter, req *http.Request) {
	results, err := s.dispatcher.HandleIncomingRequest(req)
	if err != nil {
		logrus.WithError(err).Error("could not handle incoming request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	deadline := time.After(time.Second * 10)
	for {
		select {
		case res, more := <-results:
			if !more {
				w.WriteHeader(http.StatusAccepted)
				return
			}

			if res.Err != nil {
				logrus.WithError(err).WithField("handler", res.HandlerName).Error("handler error")
			} else {
				logrus.WithFields(logrus.Fields{
					"handler":  res.HandlerName,
					"duration": res.Duration.String(),
				}).Debug("handler succeeded")
			}
		case <-deadline:
			logrus.Error("timed out while waiting for dispatcher results")
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}
}
