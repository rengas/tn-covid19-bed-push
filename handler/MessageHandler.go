package handler

import (
	"github.com/jmoiron/sqlx"
	"github.com/rengas/tn-covid19-bed-alert/model"
	"log"
	"net/http"
	"strconv"
)

type MessageHandle struct {
	Database *sqlx.DB
}

func (h MessageHandle) ViewMessageHandler(w http.ResponseWriter, r *http.Request) {
	getParams := make(map[string][]string)
	if r.ParseForm() == nil {
		for k, v := range r.Form {
			if len(v) > 0 {
				getParams[k] = v
			}
		}
	}
	action := getParams["id"]
	if len(action) > 0 {

		id, err := strconv.ParseInt(action[0], 10, 64)
		if err != nil {
			w.Write([]byte("something went wrong"))
		}
		pLogs := h.GetMessageBody(id)
		if len(pLogs) > 0 {
			//send back response
			w.Write([]byte(pLogs[0].Page))
		}

	}

}

func (h MessageHandle) GetMessageBody(id int64) []model.PushLogs {
	rows, err := h.Database.Queryx("SELECT * FROM push_logs where id=$1", id)
	if err != nil {
		log.Printf("unable to get districts %s", err.Error())
		return nil
	}
	defer rows.Close()
	pLog := make([]model.PushLogs, 0)

	for rows.Next() {
		var stat = &model.PushLogs{}
		err := rows.StructScan(stat)
		if err != nil {
			log.Println(err)
			continue
		}
		pLog = append(pLog, *stat)
	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return pLog
}
