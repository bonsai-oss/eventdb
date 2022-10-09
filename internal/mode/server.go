package mode

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/bonsai-oss/eventdb/internal/database"
	"github.com/bonsai-oss/eventdb/internal/database/model"
	"github.com/bonsai-oss/eventdb/internal/handler"
	"github.com/bonsai-oss/eventdb/internal/middleware"
)

type Server struct {
	Database      database.Settings
	Instance      http.Server
	WorkerInput   chan model.Event
	WorkerOutput  chan error
	Logger        *log.Logger
	ListenAddress string
}

func (s *Server) Run(c *kingpin.ParseContext) error {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	// initialize custom logger
	s.Logger = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)

	// initialize
	if err := s.Database.InitializeDB(s.Logger); err != nil {
		panic(err)
	}

	s.WorkerInput = make(chan model.Event)
	s.WorkerOutput = make(chan error)

	// define global router
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	// define api router
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.Logging(s.Logger))

	apiV1Router := apiRouter.PathPrefix("/v1").Subrouter()
	apiV1Router.Path("/streams/{streamName}").Methods(http.MethodPost).HandlerFunc(handler.CreateHandler(s.WorkerInput, s.WorkerOutput))
	apiV1Router.Path("/streams/{streamName}").Methods(http.MethodGet).HandlerFunc(handler.PollHandler(s.Database.Client))
	apiV1Router.Path("/event/{eventID}").Methods(http.MethodGet).HandlerFunc(handler.PollHandler(s.Database.Client))

	s.Instance = http.Server{Handler: router, Addr: s.ListenAddress}

	workerDone := make(chan bool)
	ctx, cancel := context.WithCancel(context.Background())
	go s.createWorker(ctx, workerDone)

	go func() {
		err := s.Instance.ListenAndServe()
		if err != nil {
			s.Logger.Println(err)
		}
	}()

	// wait for os interrupt
	<-sig

	// shutdown webserver
	err := s.Instance.Shutdown(context.Background())
	if err != nil {
		s.Logger.Println(err)
	}

	// stop the worker
	cancel()
	<-workerDone
	fmt.Println("goodby")

	return nil
}

func (s *Server) createWorker(ctx context.Context, done chan<- bool) {
	for {
		select {
		case <-ctx.Done():
			done <- true
			return
		case event := <-s.WorkerInput:
			s.WorkerOutput <- s.Database.Client.Create(&event).Error
		}
	}
}
