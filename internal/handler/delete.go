package handler

import (
	"net/http"

	"github.com/bonsai-oss/jsonstatus"
	"github.com/bonsai-oss/mux"
	"gorm.io/gorm"

	"github.com/bonsai-oss/eventdb/internal/database/model"
)

func DropHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		streamName, streamPresent := vars["streamName"]
		if !streamPresent {
			jsonstatus.Status{Code: http.StatusInternalServerError, Message: "going to hell!!!!"}.Encode(w)
			return
		}

		transactionError := db.Transaction(func(tx *gorm.DB) error {
			query := tx.Delete(&model.Event{}, "stream_name = ?", streamName)

			if query.Error != nil {
				return query.Error
			}

			return nil
		})

		if transactionError != nil {
			jsonstatus.Status{Code: http.StatusInternalServerError, Message: transactionError.Error()}.Encode(w)
			return
		}

		jsonstatus.Status{Code: http.StatusOK, Message: "deleted"}.Encode(w)
	}
}
