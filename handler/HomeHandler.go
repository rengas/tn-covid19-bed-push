package handler

import (
	"github.com/jmoiron/sqlx"
	"github.com/rengas/tn-covid19-bed-alert/model"
	"html/template"
	"log"
	"net/http"
)

type HomeHandle struct {
	Tmpl *template.Template
	Database *sqlx.DB
}

func (h HomeHandle) HomeHandler(w http.ResponseWriter, r *http.Request) {
		// fetch all unique districts from database
		districts := h.GetDistricts()
		if len(districts)==0 {
			return
		}
		h.Tmpl.Execute(w,districts)  // merge.
}

func (h HomeHandle) GetDistricts()map[string][]model.HospitalStatusSQL {
	rows, err := h.Database.Queryx("SELECT * FROM hospital_status")
	if err != nil {
		log.Printf("unable to get districts %s",err.Error())
		return nil
	}
	defer rows.Close()
	districts := make( map[string][]model.HospitalStatusSQL,0)

	for rows.Next() {
		var stat =&model.HospitalStatusSQL{}
		err := rows.StructScan(stat)
		if err != nil {
			log.Println(err)
			continue
		}
		if val,ok:=districts[stat.District];ok{
			val = append(val,*stat)
			districts[stat.District] =val
		}else{
			firstStat := make([]model.HospitalStatusSQL,0)
			firstStat = append(firstStat,*stat)
			districts[stat.District] =firstStat
		}

	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return districts
}