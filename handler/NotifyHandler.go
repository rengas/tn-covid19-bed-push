package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"firebase.google.com/go/messaging"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/rengas/tn-covid19-bed-alert/model"
	"html/template"
	"log"
	"math"
	"net/http"
	"errors"
)

const (
	defaultLimit = 100
	Unsent ="UNSENT"
)

type NotifyHandle struct {
	Tmpl *template.Template
	Database *sqlx.DB
	Fcm *messaging.Client
	Env string
}
func (h NotifyHandle) NotifyHandler(w http.ResponseWriter, r *http.Request) {

	go func() {
		log.Println("Notify Started...")
		hospitalStatus := h.evaluateHospitalStatus()
		total :=h.countPushSubscriptions()

		if total>0{
			limit := int64(defaultLimit)
			noPages := 0

			if total> defaultLimit {
				limit := int64(defaultLimit)
				noPages = int(math.Round(float64(total)/float64(limit)))
			}

			if noPages>0{
				for i :=0; i<noPages;i++{
					records := h.getPushSubscriptions(limit,int64(i))
					if len(records)>0{
						for i:=0;i<len(records);i++{
							pushlog,msg := h.getNotificationMessage(records[i],hospitalStatus)
							if pushlog!=nil&&msg!=nil{
								m,err := h.sendPushNotification(msg)
								if err!=nil{
									log.Printf("push not send %s", err.Error())
									continue
								}
								//update status
								pushlog.Status =m
								//update database
								h.UpdatePushLogs(*pushlog.Id,m)

							}
						}

					}
				}
			}
			records := h.getPushSubscriptions(limit,int64(noPages))
			if len(records)>0{
				for i:=0;i<len(records);i++{
					pushlog,msg := h.getNotificationMessage(records[i],hospitalStatus)
					if pushlog!=nil&&msg!=nil{
						m,err := h.sendPushNotification(msg)
						if err!=nil{
							log.Printf("push not send %s", err.Error())
							continue
						}
						//update status
						pushlog.Status =m
						//update database
						h.UpdatePushLogs(*pushlog.Id,m)

					}
				}
			}
		}
		log.Println("Notify Ended...")
	}()

}

func (h NotifyHandle) sendPushNotification(message *messaging.Message)(string,error){
	id, err := h.Fcm.Send(context.Background(), message)
	if err!=nil{
		log.Printf("unable to send push notification %s", err.Error())
		return "",err
	}
	if id==""{
		log.Printf("unable to send push notification %s", err.Error())
		return Unsent,errors.New("id is empty")
	}
	return id, nil
}

func (h NotifyHandle) evaluateHospitalStatus() map[int64]model.HospitalStatus {
	hstatusSQL  := h.getHospitalStatus()
	mOfStatus := make(map[int64]model.HospitalStatus,0)
	for _,status := range hstatusSQL{
		 st, err := h.parseHospitalStatus(status)
		 if err!=nil{
		 	continue
		 }
		 if st.Ventilator.Vacant>0 ||
		 	st.NonOxygenSupportedBeds.Vacant>0 ||
		 	st.OxygenSupportedBeds.Vacant>0 ||
		 	st.ICUBeds.Vacant>0 ||
		 	st.CovidBeds.Vacant>0{
			 mOfStatus[status.Id] = *st
		 }
	}
	return mOfStatus
}

func (h NotifyHandle)parseHospitalStatus(hStatus model.HospitalStatusSQL)(*model.HospitalStatus,error){

	sb,err := json.Marshal(hStatus)
	if err!=nil{
		return nil,err
	}
	var hs model.HospitalStatus

	err = json.Unmarshal(sb,&hs)
	if err!=nil{
		return nil,err
	}
	return &hs,nil
}

func (h NotifyHandle) getHospitalStatus() []model.HospitalStatusSQL {
	rows, err := h.Database.Queryx("SELECT * FROM hospital_status")
	if err != nil {
		log.Printf("unable to get hospitalStatus %s",err.Error())
		return nil
	}
	defer rows.Close()
	hospitals := make( []model.HospitalStatusSQL,0)

	for rows.Next() {
		var stat =&model.HospitalStatusSQL{}
		err := rows.StructScan(stat)
		if err != nil {
			log.Println(err)
			continue
		}
		hospitals = append(hospitals,*stat)
	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil
	}
	return hospitals
}

func (h NotifyHandle) countPushSubscriptions() int64{
		row := h.Database.QueryRow("SELECT count(*) FROM push_subscription")
		var count int64
		err := row.Scan(&count)
		if err != nil {
			log.Printf("unable to get count %s",err.Error())
			return 0
		}

		return  count

}

func  (h NotifyHandle) getPushSubscriptions(limit int64,offset int64) []model.UserSubscription {
	rows, err := h.Database.Queryx("SELECT * FROM push_subscription limit $1 offset $2",limit,offset)
	if err != nil {
		log.Printf("unable to get push subscriptions %s",err.Error())
		return nil
	}
	defer rows.Close()
	subscriptions := make( []model.UserSubscription,0)

	for rows.Next() {
		var sub =&model.UserSubscription{}
		err := rows.StructScan(sub)
		if err != nil {
			log.Println(err)
			continue
		}
		subscriptions = append(subscriptions,*sub)

	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		log.Printf("unable to get rows %s",err.Error())
	}
	return subscriptions
}

func (h NotifyHandle) CreatePushLogs(log *model.PushLogs) (int64,error){
	sqlStatement := `
			INSERT INTO public.push_logs (title, body,page,status)
			VALUES ($1,$2,$3,$4) RETURNING id`
	var id int64
	err := h.Database.Get(&id,sqlStatement,
		log.Title,
		log.Body,
		log.Page,
		log.Status,
	)
	if err!=nil{
		return 0,err
	}
	return id,err
}

func (h NotifyHandle) UpdatePushLogs(id int64, status string){
	sqlStatement := `UPDATE public.push_logs SET status =$1 where id=$2;`
	_, err := h.Database.Exec(sqlStatement,status,id)
	if err != nil {
		log.Printf("unable to update push log %s", err.Error())
	}
}

func (h NotifyHandle) getNotificationMessage(u model.UserSubscription, hs map[int64]model.HospitalStatus)(*model.PushLogs,*messaging.Message){
	str :=
		"<!DOCTYPE html>\n<html>\n<head>\n<style>\ntable {\n  font-family: arial, sans-serif;\n  border-collapse: collapse;\n  width: 100%;\n}\n\ntd, th {\n  border: 1px solid #dddddd;\n  text-align: left;\n  padding: 8px;\n}\n\ntr:nth-child(even) {\n  background-color: #dddddd;\n}\n</style>\n</head>\n<body>\n\n" +
			"\n<h2>Available Beds & Ventilators</h2>\n" +
			"<h2> Get latest update here </h2>\n" +
			"<a  href=\"https://stopcorona.tn.gov.in/beds.php\">https://stopcorona.tn.gov.in/beds.php</a>" +
			"<table>  <tr>" +
			"<th>District</th>" +
			"<th>Hospital</th> " +
			"<th>Available Covid beds</th>" +
			"<th>Available Oxygen Supported beds</th>" +
			"<th>Available Non Oxygen  beds</th>" +
			"<th>Available ICU beds</th>" +
			"<th>Available Ventilator</th>" +
			"<th>Last Updated By hospital</th>" +
			"<th>Contact</th>" +
			"<th>Remarks</th>" +
			"<th>Last Synced</th>" +
			"</tr>"
	covidVacant := int64(0)
	oxygenBeds := int64(0)
	nonOxygenBeds :=int64(0)
	icuVacant :=int64(0)
	ventilator :=int64(0)

	for _,subs:= range u.Subscriptions{
		if v,ok := hs[subs];ok{
			covidVacant = v.CovidBeds.Vacant + covidVacant
			oxygenBeds = v.OxygenSupportedBeds.Vacant + oxygenBeds
			nonOxygenBeds = v.NonOxygenSupportedBeds.Vacant + nonOxygenBeds
			icuVacant = v.ICUBeds.Vacant + icuVacant
			ventilator = v.Ventilator.Vacant + ventilator

			var buf bytes.Buffer
			err := h.Tmpl.Execute(&buf,v)
			if err!=nil{
				continue
			}
			str = str+buf.String()
		}
	}
	str =str+ "</table>"+
		"</body>\n</html>"

	title := "TN Hospital bed status"
	body := fmt.Sprintf("Covid : %d \n" +
		"O: %d \n" +
		"Non-O2 : %d \n" +
		"ICU : %d \n" +
		"Ventilator: %d \n",
		covidVacant,
		oxygenBeds,
		nonOxygenBeds,
		icuVacant,
		ventilator,
	)

	pushLog := &model.PushLogs{
		Title:  title,
		Body:   body,
		Page:   str,
		Status: Unsent,
	}
	id,err := h.CreatePushLogs(pushLog)
	if err!=nil{
		log.Printf("unable to create push log %s",err.Error())
		return nil,nil
	}

	pushLog.Id = &id

	link := fmt.Sprintf("http://localhost:8000/message?id=%d",*pushLog.Id)
	if h.Env=="prd"{
		//TODO this should be replaced with actual domain
		link=fmt.Sprintf("https://tn-covid-beds.renga.me/message?id=%d",id)
	}

	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Webpush: &messaging.WebpushConfig{
			FcmOptions: &messaging.WebpushFcmOptions{
				Link: link,
			},
		},
		Token: u.Token,

	}

	return pushLog,message
}