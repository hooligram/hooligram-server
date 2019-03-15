package main

import (
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

func main() {
	defer db.Close()

	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("[MAIN] PORT must be set. Exiting...")
	}

	twilioAPIKey = os.Getenv("TWILIO_API_KEY")

	if twilioAPIKey == "" {
		log.Fatal("[MAIN] TWILIO_API_KEY must be set. Exiting...")
	}

	if db == nil {
		log.Fatal("[MAIN] Failed to initialize the DB. Exiting...")
	}

	router := mux.NewRouter()
	router.HandleFunc("/v2", v2)

	go broadcast()
	go deliverMessage()

	http.ListenAndServe(":"+port, router)
}

func broadcast() {
	for {
		action := <-broadcastChan
		verifiedClients := findAllVerifiedClients()

		for _, verifiedClient := range verifiedClients {
			var onlineConn *websocket.Conn

			for conn, client := range clients {
				if client.CountryCode == verifiedClient.CountryCode &&
					client.PhoneNumber == verifiedClient.PhoneNumber &&
					client.IsSignedIn {
					onlineConn = conn
				}
			}

			if onlineConn == nil {
				if _, ok := pendingActionQueue[verifiedClient]; !ok {
					pendingActionQueue[verifiedClient] = []*Action{}
				}

				pendingActions := pendingActionQueue[verifiedClient]
				pendingActions = append(pendingActions, action)
				pendingActionQueue[verifiedClient] = pendingActions
				continue
			}

			onlineConn.WriteJSON(action)
		}
	}
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
