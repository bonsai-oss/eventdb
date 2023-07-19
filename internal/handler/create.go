package handler

import (
	"encoding/json"
	"net/http"

	"github.com/bonsai-oss/jsonstatus"
	"github.com/bonsai-oss/mux"

	"github.com/bonsai-oss/eventdb/v2/internal/database/model"
)

func CreateHandler(workerInput chan<- model.Event, workerOutput <-chan error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		event := model.Event{}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		decoder.UseNumber()

		switch r.Header.Get("Content-Type") {
		case "application/vnd+eventdb.event+json":
			transferEvent := model.TransferEvent{}
			if err := decoder.Decode(&transferEvent); err != nil {
				jsonstatus.Status{Code: http.StatusUnprocessableEntity, Message: err.Error()}.Encode(w)
				return
			}

			event.Data = transferEvent.Data
			event.Type = transferEvent.Type
		case "application/json":
			if r.Header.Get("X-Event-Type") == "" {
				jsonstatus.Status{Code: http.StatusUnprocessableEntity, Message: "missing event type"}.Encode(w)
				return
			}
			data := make(map[string]interface{})
			if err := decoder.Decode(&data); err != nil {
				jsonstatus.Status{Code: http.StatusUnprocessableEntity, Message: err.Error()}.Encode(w)
				return
			}
			event.Data = data
			event.Type = r.Header.Get("X-Event-Type")
		default:
			jsonstatus.Status{Code: http.StatusUnsupportedMediaType, Message: "unsupported content type"}.Encode(w)
			return
		}

		event.StreamName = vars["streamName"]
		workerInput <- event

		databaseError := <-workerOutput
		if databaseError != nil {
			jsonstatus.Status{Code: http.StatusInternalServerError, Message: databaseError.Error()}.Encode(w)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
