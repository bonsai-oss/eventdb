package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bonsai-oss/jsonstatus"
	"github.com/bonsai-oss/mux"
	cloudevents "github.com/cloudevents/sdk-go/v2"

	"github.com/bonsai-oss/eventdb/v2/internal/database/model"
)

func CloudEventsCreateHandler(workerInput chan<- model.Event, workerOutput <-chan error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var event model.Event

		ev, err := cloudevents.NewEventFromHTTPRequest(r)
		if err != nil {
			jsonstatus.Status{Code: http.StatusBadRequest, Message: err.Error()}.Encode(w)
			return
		}

		if ev.DataContentType() != "application/json" {
			log.Printf("invalid content type: %s", ev.DataContentType())
			jsonstatus.Status{Code: http.StatusUnsupportedMediaType, Message: fmt.Sprintf("unsupported content type %v", ev.DataContentType())}.Encode(w)
			return
		}

		event.Type = ev.Type()

		event.StreamName = vars["streamName"]
		dataUnmarshalError := json.Unmarshal(ev.Data(), &event.Data)
		if dataUnmarshalError != nil {
			jsonstatus.Status{Code: http.StatusUnprocessableEntity, Message: dataUnmarshalError.Error()}.Encode(w)
			return
		}

		workerInput <- event

		databaseError := <-workerOutput
		if databaseError != nil {
			jsonstatus.Status{Code: http.StatusInternalServerError, Message: databaseError.Error()}.Encode(w)
			return
		}

		jsonstatus.Status{Code: http.StatusCreated, Message: ev.String()}.Encode(w)
	}
}
