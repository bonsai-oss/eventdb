package handler

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"golang.fsrv.services/eventdb/internal/database/model"
	"golang.fsrv.services/jsonstatus"
	"gorm.io/gorm"
	"log"
	"net/http"
	"sort"
)

func PollHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		streamName, streamPresent := vars["streamName"]
		eventID, eventIDPresent := vars["eventID"]
		if streamPresent && eventIDPresent {
			jsonstatus.Status{StatusCode: http.StatusInternalServerError, Message: "going to hell!!!!"}.Encode(w)
			return
		}
		view := model.EventView{}
		var events []model.Event

		query := db.Find(&model.Event{})
		if streamPresent {
			query = query.Where("stream_name = ?", streamName)
		} else if eventIDPresent {
			query = query.Where("id = ?", eventID)
		}
		tx := query.Find(&events)
		view.Entries = append(view.Entries, events...)

		w.Header().Set("Content-Type", "application/json")
		switch tx.Error {
		case nil:
			// backward sort events by updated_at time
			if len(view.Entries) > 0 {
				sort.Slice(view.Entries, func(j, i int) bool {
					return view.Entries[i].UpdatedAt.Before(view.Entries[j].UpdatedAt)
				})
				view.LastModified = view.Entries[0].UpdatedAt
			}

			// encode view to client
			if err := json.NewEncoder(w).Encode(view); err != nil {
				log.Println(err)
			}
		case gorm.ErrRecordNotFound:
			jsonstatus.Status{StatusCode: http.StatusNotFound, Message: tx.Error.Error()}.Encode(w)
		default:
			jsonstatus.Status{StatusCode: http.StatusInternalServerError, Message: tx.Error.Error()}.Encode(w)
		}
	}
}
