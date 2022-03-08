package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"golang.fsrv.services/eventdb/internal/database/model"
	"golang.fsrv.services/jsonstatus"
)

func CreateHandler(workerInput chan<- model.Event, workerOutput <-chan error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		transferEvent := model.TransferEvent{}
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		decoder.UseNumber()
		err := decoder.Decode(&transferEvent)
		if err != nil {
			jsonstatus.Status{StatusCode: http.StatusUnprocessableEntity, Message: err.Error()}.Encode(w)
			return
		}

		event := model.Event{}
		event.Data = transferEvent.Data
		event.Type = transferEvent.Type
		event.StreamName = vars["streamName"]

		workerInput <- event

		databaseError := <-workerOutput
		if databaseError != nil {
			jsonstatus.Status{StatusCode: http.StatusInternalServerError, Message: err.Error()}.Encode(w)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}
