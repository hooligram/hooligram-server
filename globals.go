package main

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]*Client)
var db *sql.DB
var httpClient = &http.Client{}
var messageDeliveryChan = make(chan *MessageDelivery)
var pendingActionQueue = make(map[*Client][]*Action)
var twilioAPIKey string
var upgrader = websocket.Upgrader{}

func getClient(conn *websocket.Conn) (*Client, error) {
	client, ok := clients[conn]

	if !ok {
		return nil, errors.New("i couldn't find you")
	}

	return client, nil
}

func getSignedInClient(conn *websocket.Conn) (*Client, error) {
	client, err := getClient(conn)

	if err != nil {
		return nil, err
	}

	if !client.IsSignedIn {
		return nil, errors.New("you need to sign in first")
	}

	return client, nil
}

func signIn(
	conn *websocket.Conn,
	countryCode, phoneNumber, verificationCode string,
) (*Client, error) {
	client, ok := findVerifiedClient(countryCode, phoneNumber, verificationCode)

	if !ok {
		return nil, errors.New("couldn't find such verified client")
	}

	clients[conn].ID = client.ID
	clients[conn].CountryCode = client.CountryCode
	clients[conn].PhoneNumber = client.PhoneNumber
	clients[conn].VerificationCode = verificationCode
	clients[conn].DateCreated = client.DateCreated
	clients[conn].IsSignedIn = true

	return clients[conn], nil
}

func signOut(conn *websocket.Conn) {
	delete(clients, conn)
}

func unverifyClient(client *Client, conn *websocket.Conn) error {
	return verifyClient(client, conn, "")
}

func verifyClient(client *Client, conn *websocket.Conn, verificationCode string) error {
	err := updateClientVerificationCode(client, verificationCode)

	if err != nil {
		delete(clients, conn)
		return err
	}

	clients[conn].VerificationCode = verificationCode

	return nil
}

func writeQueuedActions(client *Client) {
	for pendingClient := range pendingActionQueue {
		countryCodeMatch := pendingClient.CountryCode == client.CountryCode
		phoneNumberMatch := pendingClient.PhoneNumber == client.PhoneNumber
		isCurrentClient := countryCodeMatch && phoneNumberMatch

		if isCurrentClient {
			queueLen := len(pendingActionQueue[pendingClient])
			for i := queueLen - 1; i >= 0; i-- {
				client.writeJSON(pendingActionQueue[pendingClient][i])
			}

			delete(pendingActionQueue, pendingClient)
		}
	}
}
