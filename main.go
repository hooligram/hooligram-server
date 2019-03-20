package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

const mainTag = "main"

func main() {
	defer db.Close()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal(fmt.Sprintf("[%v] PORT not set. exiting...", mainTag))
	}

	twilioAPIKey = os.Getenv("TWILIO_API_KEY")
	if twilioAPIKey == "" {
		log.Fatal(fmt.Sprintf("[%v] TWILIO_API_KEY not set. exiting...", mainTag))
	}

	if db == nil {
		log.Fatal(fmt.Sprintf("[%v] db not found. exiting...", mainTag))
	}

	router := mux.NewRouter()
	router.HandleFunc("/v2", v2)

	go deliverMessage()

	http.ListenAndServe(":"+port, router)
}

func deliverMessage() {
	for {
		messageDelivery := <-messageDeliveryChan
		message := messageDelivery.Message
		recipientIDs := messageDelivery.RecipientIDs

		for conn, client := range clients {
			if !containsID(recipientIDs, client.ID) {
				continue
			}

			if conn == nil {
				continue
			}

			action := constructDeliverMessageAction(message)
			conn.WriteJSON(action)
		}
	}
}
