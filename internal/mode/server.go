package mode

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.fsrv.services/eventdb/internal/database"
	"golang.fsrv.services/eventdb/internal/database/model"
	"golang.fsrv.services/eventdb/internal/handler"
	"golang.fsrv.services/eventdb/internal/middleware"
)

type Server struct {
	Database      database.Settings
	Instance      http.Server
	WorkerInput   chan model.Event
	WorkerOutput  chan error
	Logger        *log.Logger
	ListenAddress string
}

func (s *Server) Initialize() {
	// parse command line parameters
	flag.StringVar(&s.Database.Database, "database.name", "eventdb", "name of the database")
	flag.StringVar(&s.Database.Username, "database.user", "postgres", "username of the database")
	flag.StringVar(&s.Database.Password, "database.password", "test123", "password of the database")
	flag.StringVar(&s.Database.Host, "database.host", "", "address of the database")
	flag.StringVar(&s.ListenAddress, "web.listen-address", ":8080", "address listening on")
	flag.Parse()

	if e := os.Getenv("DATABASE_NAME"); e != "" {
		s.Database.Database = e
	}
	if e := os.Getenv("DATABASE_USER"); e != "" {
		s.Database.Username = e
	}
	if e := os.Getenv("DATABASE_PASSWORD"); e != "" {
		s.Database.Password = e
	}
	if e := os.Getenv("DATABASE_HOST"); e != "" {
		s.Database.Host = e
	}

	// initialize custom logger
	s.Logger = log.New(os.Stdout, "", log.Ltime|log.Lshortfile)

	// initialize
	if err := s.Database.InitializeDB(s.Logger); err != nil {
		panic(err)
	}

	s.WorkerInput = make(chan model.Event)
	s.WorkerOutput = make(chan error)
}

func (s *Server) Run(sig <-chan os.Signal) {
	// define global router
	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)

	// define api router
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(middleware.Logging(s.Logger))

	apiV1Router := apiRouter.PathPrefix("/v1").Subrouter()
	apiV1Router.HandleFunc("/streams/{streamName}", handler.CreateHandler(s.WorkerInput, s.WorkerOutput)).Methods(http.MethodPost)
	apiV1Router.HandleFunc("/streams/{streamName}", handler.PollHandler(s.Database.Client)).Methods(http.MethodGet)
	apiV1Router.HandleFunc("/event/{eventID}", handler.PollHandler(s.Database.Client)).Methods(http.MethodGet)

	s.Instance = http.Server{Handler: router, Addr: "127.0.0.1:8080"}

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
