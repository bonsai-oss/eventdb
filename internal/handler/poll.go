package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"

	"github.com/bonsai-oss/jsonstatus"
	"github.com/bonsai-oss/mux"
	"gorm.io/gorm"

	"github.com/bonsai-oss/eventdb/v2/internal/database/model"
)

func PollHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		streamName, streamPresent := vars["streamName"]
		eventID, eventIDPresent := vars["eventID"]
		if streamPresent && eventIDPresent {
			jsonstatus.Status{Code: http.StatusInternalServerError, Message: "going to hell!!!!"}.Encode(w)
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
			jsonstatus.Status{Code: http.StatusNotFound, Message: tx.Error.Error()}.Encode(w)
		default:
			jsonstatus.Status{Code: http.StatusInternalServerError, Message: tx.Error.Error()}.Encode(w)
		}
	}
}
