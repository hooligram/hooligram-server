package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
)

var broadcastChan = make(chan *Action)
var clients = make(map[*websocket.Conn]*Client)
var db *sql.DB
var httpClient = &http.Client{}
var pendingActionQueue = make(map[*Client][]*Action)
var twilioAPIKey string
var upgrader = websocket.Upgrader{}

func signIn(
	conn *websocket.Conn,
	countryCode, phoneNumber, verificationCode string,
) (*Client, error) {
	client, ok := findVerifiedClient(countryCode, phoneNumber, verificationCode)

	if !ok {
		return nil, errors.New("couldn't find such verified client")
	}

	client.conn = conn
	client.IsSignedIn = true
	clients[conn] = client

	return clients[conn], nil
}

func signOut(conn *websocket.Conn) {
	delete(clients, conn)
}

func writeQueuedActions(client *Client) {
	for pendingClient := range pendingActionQueue {
		countryCodeMatch := pendingClient.CountryCode == client.CountryCode
		phoneNumberMatch := pendingClient.PhoneNumber == client.PhoneNumber
		isCurrentClient := countryCodeMatch && phoneNumberMatch

		if isCurrentClient {
			for _, pendingAction := range pendingActionQueue[pendingClient] {
				client.writeJSON(pendingAction)
			}

			delete(pendingActionQueue, pendingClient)
		}
	}
}
