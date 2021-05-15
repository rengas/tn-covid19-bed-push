package handler

import (
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rengas/tn-covid19-bed-alert/model"
	"log"
	"net/http"
	"strconv"
)
type SubHandle struct {
	Database *sqlx.DB
}

func (s SubHandle)  SubscribeHandler(w http.ResponseWriter, r *http.Request) {
	// an example API handler
	r.ParseForm()
	subscription := &model.UserSubscription{}
	for k, v := range r.Form {
		if k=="email"{
			subscription.Email =v[0]
			continue
		}
		if k=="token"{
			subscription.Token =v[0]
		}
		if k!="district"{
			for _,id := range v{
				subId, err := strconv.ParseInt(id,10,64)
				if err!=nil{
					continue
				}
				subscription.Subscriptions= append(subscription.Subscriptions,subId)
			}
		}
	}
	if subscription.Email!="" && len(subscription.Subscriptions)>0{
		subs := s.getSubscription(subscription.Email)
		if len(subs)>0{
			//update existings subs
			s.updateSubscription(subs,*subscription)
			//create a new sub
			s.createSubscription(*subscription)
		}else{
			s.createSubscription(*subscription)
		}
	}
	fmt.Println(subscription)
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}

func (s SubHandle)  UnSubscribeHandler(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
	subscription := &model.UserSubscription{}
	for k, v := range r.Form {
		if k == "email" {
			subscription.Email = v[0]
			continue
		}
	}
	if subscription.Email!=""{
		subscriptions := s.getSubscription(subscription.Email)
		if len(subscriptions)>0{
			err :=s.deleteSubscription(subscriptions)
			if err!=nil{
				log.Println("unable to delete subscription")
			}
		}
	}
}

func (s SubHandle) deleteSubscription(subs []model.UserSubscription) error{
	for _,sub := range subs{
		sqlStatement := `
			DELETE From  push_subscription where id =$1`
		_, err := s.Database.Exec(sqlStatement,
			sub.Id,
		)
		if err!=nil{
			return err
		}
	}
	return nil
}

func (s SubHandle) getSubscription(email string) []model.UserSubscription {
	rows, err := s.Database.Queryx("SELECT * FROM push_subscription where email=$1",email)
	if err != nil {
		log.Printf("unable to get districts %s",err.Error())
		return nil
	}
	defer rows.Close()
	subs := make( []model.UserSubscription,0)

	for rows.Next() {
		var stat =&model.UserSubscription{}
		err := rows.StructScan(stat)
		if err != nil {
			log.Println(err)
			continue
		}
		subs = append(subs,*stat)
	}
	// get any error encountered during iteration
	err = rows.Err()
	if err != nil {
		log.Println(err)
		return nil
	}
	return subs

}

func (s SubHandle) createSubscription(user model.UserSubscription)error{
		sqlStatement := `
			INSERT INTO push_subscription (email, subscriptions,token)
			VALUES ($1,$2,$3)`
		_, err := s.Database.Exec(sqlStatement,
			user.Email,
			pq.Array(user.Subscriptions),
			user.Token,
		)
		if err!=nil{
			return err
		}
		return err

}

func (s SubHandle) updateSubscription(userSub []model.UserSubscription,update model.UserSubscription)error{

	for _,sub := range userSub {

		sqlStatement := `UPDATE public.push_subscription SET email =$1,subscriptions =$2 where id=$3;`
		_, err := s.Database.Exec(sqlStatement,sub.Email, pq.Array(update.Subscriptions),update.Id)
		if err != nil {
			log.Printf("unable to update subscription %s", err.Error())
			continue
		}

	}

	return nil

}