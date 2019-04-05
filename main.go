package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/hooligram/hooligram-server/delivery"
	"github.com/hooligram/hooligram-server/notifications"
	"github.com/hooligram/hooligram-server/v2"
)

const mainTag = "main"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal(fmt.Sprintf("[%v] PORT not set", mainTag))
	}

	fcmmessaginghost := os.Getenv("FCM_MESSAGING_HOST")
	if fcmmessaginghost == "" {
		log.Fatal(fmt.Sprintf("[%v] FCM_MESSAGING_HOST not set", mainTag))
	}

	fcmmessagingauthkey := os.Getenv("FCM_MESSAGING_AUTHKEY")
	if fcmmessagingauthkey == "" {
		log.Fatal(fmt.Sprintf("[%v] FCM_MESSAGING_AUTHKEY not set", mainTag))
	}

	notifications.Init(fcmmessagingauthkey, fcmmessaginghost)

	router := mux.NewRouter()
	router.HandleFunc("/v2", v2.V2)

	go delivery.DeliverMessage()
	go delivery.DeliverMessageGroup()
	go notifications.HandleNotification()

	http.ListenAndServe(":"+port, router)
}
