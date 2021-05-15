package model

import (
	"encoding/json"
	"firebase.google.com/go/messaging"
	sqlxtypes "github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
)

type HospitalStatus struct {
	District               string `db:"district" ,json:"district"`
	Institution            string `db:"institution" ,json:"institution"`
	CovidBeds              Status `db:"covid_beds" ,json:"covidBeds"`
	OxygenSupportedBeds    Status `db:"oxygen_supported_beds" ,json:"oxygenSupportedBeds"`
	NonOxygenSupportedBeds Status `db:"non_oxygen_supported_beds" ,json:"nonOxygenSupportedBeds"`
	ICUBeds                Status `db:"icu_beds" ,json:"icuBeds"`
	Ventilator             Status `db:"ventilator" ,json:"ventilator"`
	LastUpdate             string `db:"last_updated" ,json:"last_updated"`
	ContactNumber          string `db:"contact_number" ,json:"contactNumber"`
	Remarks                string `db:"remarks" ,json:"remarks"`
	UpdatedAt              string `db:"updated_at" ,json:"UpdatedAt"`
	CreatedAt              string `db:"created_at" ,json:"createdAt"`
}

type UserSubscription struct {
	Id int64 `db:"id"`
	Email string `db:"email"`
	Subscriptions pq.Int64Array `db:"subscriptions"`
	Token string `db:"token"`
	UpdatedAt string `db:"updated_at"`
	CreatedAt string `db:"created_at"`
}

type HospitalStatusSQL struct {
	Id int64
	District string `db:"district" ,json:"district"`
	Institution string `db:"institution" ,json:"institution"`
	CovidBeds sqlxtypes.JSONText `db:"covid_beds" ,json:"covidBeds"`
	OxygenSupportedBeds sqlxtypes.JSONText `db:"oxygen_supported_beds" ,json:"oxygenSupportedBeds"`
	NonOxygenSupportedBeds sqlxtypes.JSONText`db:"non_oxygen_supported_beds" ,json:"nonOxygenSupportedBeds"`
	ICUBeds sqlxtypes.JSONText `db:"icu_beds" ,json:"icuBeds"`
	Ventilator  sqlxtypes.JSONText `db:"ventilator" ,json:"ventilator"`
	LastUpdate string `db:"last_updated" ,json:"last_updated"`
	ContactNumber string `db:"contact_number" ,json:"contactNumber"`
	Remarks string `db:"remarks" ,json:"remarks"`
	UpdatedAt string `db:"updated_at" ,json:"UpdatedAt"`
	CreatedAt string `db:"created_at" ,json:"createdAt"`
}

type PushLogs struct {
	Id *int64
	Title string
	Body string
	Page string
	Status string
}

type PushMessage struct {
	PushLogs
	messaging.Message
}


type Status struct {
	Total int64 // use string to simplify
	Occupied int64
	Vacant int64
}

func (s Status) JsonString()string{
	b, err := json.Marshal(s)
	if err!=nil{
		return ""
	}
	return string(b)
}

