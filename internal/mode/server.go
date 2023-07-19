package mode

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/alecthomas/kingpin/v2"
	"github.com/bonsai-oss/mux"
	"github.com/bonsai-oss/workering/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/bonsai-oss/eventdb/internal/database"
	"github.com/bonsai-oss/eventdb/internal/database/model"
	"github.com/bonsai-oss/eventdb/internal/handler"
	"github.com/bonsai-oss/eventdb/internal/middleware"
)

type Server struct {
	Database      database.Settings
	WorkerInput   chan model.Event
	WorkerOutput  chan error
	Logger        *log.Logger
	ListenAddress string
}

func (s *Server) Run(_ *kingpin.ParseContext) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	// initialize custom logger
	s.Logger = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)

	// initialize database
	if err := s.Database.InitializeDB(s.Logger); err != nil {
		panic(err)
	}

	s.WorkerInput = make(chan model.Event)
	s.WorkerOutput = make(chan error)

	workering.Register(
		workering.RegisterSet{Name: "create", Worker: s.createWorkerBuilder()},
		workering.RegisterSet{Name: "web", Worker: s.webListenerBuilder()},
	)

	if err := workering.StartAll(); err != nil {
		panic(err)
	}

	// wait for os interrupt
	<-sig

	return workering.StopAll()
}

func (s *Server) webListenerBuilder() workering.WorkerFunction {
	return func(ctx context.Context, done chan<- any) {
		defer func() { done <- true }()
		// define global router
		router := mux.NewRouter()
		router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

		// define api router
		apiRouter := router.PathPrefix("/api").Subrouter()
		apiRouter.Use(middleware.Logging(s.Logger))

		apiV1Router := apiRouter.PathPrefix("/v1").Subrouter()
		apiV1Router.Path("/streams/{streamName}/drop").Methods(http.MethodPost).HandlerFunc(handler.DropHandler(s.Database.Client))
		apiV1Router.Path("/streams/{streamName}").Methods(http.MethodPost).HandlerFunc(handler.CreateHandler(s.WorkerInput, s.WorkerOutput))
		apiV1Router.Path("/streams/{streamName}").Methods(http.MethodGet).HandlerFunc(handler.PollHandler(s.Database.Client))
		apiV1Router.Path("/event/{eventID}").Methods(http.MethodGet).HandlerFunc(handler.PollHandler(s.Database.Client))

		httpServer := http.Server{Handler: router, Addr: s.ListenAddress}

		go func() {
			if err := httpServer.ListenAndServe(); err != nil {
				s.Logger.Println(err)
				return
			}
		}()

		<-ctx.Done()
		if err := httpServer.Shutdown(context.Background()); err != nil {
			s.Logger.Println(err)
			return
		}
	}
}

func (s *Server) createWorkerBuilder() workering.WorkerFunction {
	return func(ctx context.Context, done chan<- any) {
		defer func() { done <- true }()
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-s.WorkerInput:
				s.WorkerOutput <- s.Database.Client.Create(&event).Error
			}
		}
	}
}
