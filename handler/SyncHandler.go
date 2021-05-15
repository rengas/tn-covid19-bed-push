package handler

import (
	"github.com/gocolly/colly"
	"github.com/jmoiron/sqlx"
	"github.com/rengas/tn-covid19-bed-alert/model"
	"log"
	"net/http"
	"strconv"
)

type SyncHandle struct {
	Database *sqlx.DB
}

func (s SyncHandle) SyncHandler(w http.ResponseWriter, r *http.Request) {
	getParams := make(map[string][]string)
	if r.ParseForm() == nil {
		for k, v := range r.Form {
			if len(v) > 0 {
				getParams[k] = v
			}
		}
	}

	c := colly.NewCollector(
		colly.AllowedDomains("stopcorona.tn.gov.in","www.stopcorona.tn.gov.in"),
	)
	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("visiting", r.URL.String())
	})
	HospitatStat := make([]*model.HospitalStatus,0)
	c.OnHTML(`table[id=dtBasicExample]`, func(e *colly.HTMLElement) {
		// Iterate over rows of the table which contains different information
		// about the course

		e.ForEach("tr", func(_ int, el *colly.HTMLElement) {
			if el.ChildText("td:nth-child(2)")!=""{
				sta := &model.HospitalStatus{
					District : el.ChildText("td:nth-child(2)"),
					Institution:  el.ChildText("td:nth-child(3)"),
					CovidBeds: model.Status{
						Total:    getInt64(el.ChildText("td:nth-child(4)")),
						Occupied: getInt64(el.ChildText("td:nth-child(5)")),
						Vacant:   getInt64(el.ChildText("td:nth-child(6)")),
					},
					OxygenSupportedBeds: model.Status{
						Total:    getInt64(el.ChildText("td:nth-child(7)")),
						Occupied: getInt64(el.ChildText("td:nth-child(8)")),
						Vacant:   getInt64(el.ChildText("td:nth-child(9)")),
					},
					NonOxygenSupportedBeds: model.Status{
						Total:    getInt64(el.ChildText("td:nth-child(10)")),
						Occupied: getInt64(el.ChildText("td:nth-child(11)")),
						Vacant:   getInt64(el.ChildText("td:nth-child(12)")),
					},
					ICUBeds: model.Status{
						Total:    getInt64(el.ChildText("td:nth-child(13)")),
						Occupied: getInt64(el.ChildText("td:nth-child(14)")),
						Vacant:   getInt64(el.ChildText("td:nth-child(15)")),
					},
					Ventilator: model.Status{
						Total:    getInt64(el.ChildText("td:nth-child(16)")),
						Occupied: getInt64(el.ChildText("td:nth-child(17)")),
						Vacant:   getInt64(el.ChildText("td:nth-child(18)")),
					},
					LastUpdate: el.ChildText("td:nth-child(19)"),
					ContactNumber: el.ChildText("td:nth-child(20)"),
					Remarks:el.ChildText("td:nth-child(21)"),
				}
				HospitatStat = append(HospitatStat,sta)
			}
		})

	})


	// Start scraping on http://coursera.com/browse
	c.Visit("https://stopcorona.tn.gov.in/beds.php")
	// Print to check the slice's content

	action:= getParams["action"]
	if len(action)>0{
		if action[0] =="update"{
			s.updateHospitaStatus(HospitatStat)
		}
		if action[0]=="create"{
			s.insertHospitalStatus(HospitatStat)
		}
	}


}

func getInt64 (num string)int64{
	c, err := strconv.ParseInt(num,10,64)
	if err!=nil{
		return 0
	}
	return c
}

func (s SyncHandle) insertHospitalStatus(hStatus []*model.HospitalStatus){
	for _, st := range hStatus{
		sqlStatement := `
			INSERT INTO hospital_status (district, institution, covid_beds,
				oxygen_supported_beds,non_oxygen_supported_beds,
				icu_beds,ventilator,last_updated,contact_number,
				remarks
				)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
		_, err := s.Database.Exec(sqlStatement,
			st.District,
			st.Institution,
			st.CovidBeds.JsonString(),
			st.OxygenSupportedBeds.JsonString(),
			st.NonOxygenSupportedBeds.JsonString(),
			st.ICUBeds.JsonString(),
			st.Ventilator.JsonString(),
			st.LastUpdate,
			st.ContactNumber,
			st.Remarks,
			)
		if err != nil {
			log.Printf("unable to insert %s",err.Error())
			continue

		}
		s.Database.Exec(sqlStatement)
	}
}

func (s SyncHandle) updateHospitaStatus(hStatus []*model.HospitalStatus){

	for _, st := range hStatus{
		sqlStatement := `
			UPDATE hospital_status 
			SET
			district =$1 , institution=$2, covid_beds=$3,
				oxygen_supported_beds=$4,non_oxygen_supported_beds=$5,
				icu_beds=$6,ventilator=$7,last_updated=$8,contact_number=$9,
				remarks=$10
			where district =$11  and institution=$12;`
		_, err := s.Database.Exec(sqlStatement,
			st.District,
			st.Institution,
			st.CovidBeds.JsonString(),
			st.OxygenSupportedBeds.JsonString(),
			st.NonOxygenSupportedBeds.JsonString(),
			st.ICUBeds.JsonString(),
			st.Ventilator.JsonString(),
			st.LastUpdate,
			st.ContactNumber,
			st.Remarks,
			st.District,
			st.Institution,
		)
		if err != nil {
			log.Printf("unable to update records %s",err.Error())
			continue

		}
		s.Database.Exec(sqlStatement)
	}

}