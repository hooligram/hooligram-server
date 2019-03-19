package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
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

	go broadcast()
	go deliverMessage()

	http.ListenAndServe(":"+port, router)
}

func broadcast() {
	for {
		action := <-broadcastChan
		verifiedClients, err := findAllVerifiedClients()
		if err != nil {
			logInfo(mainTag, "error finding verified clients. "+err.Error())
			continue
		}

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
