package client

import (
	"context"
	"firebase.google.com/go/messaging"
	"log"
)


type FCMClient struct {
	Fcm *messaging.Client
}

func (h FCMClient) SendPushNotification(message *messaging.Message)(string,error){
	id, err := h.Fcm.Send(context.Background(), message)
	if err!=nil{
		log.Printf("unable to send push notification %s", err.Error())
		return "",err
	}
	return id, nil
}