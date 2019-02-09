package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

func v1(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("[V1] Failed to upgrade to WebSocket connection.")
		return
	}

	clients[conn] = &Client{}
	defer delete(clients, conn)
	defer conn.Close()

	for {
		action := Action{}
		err = conn.ReadJSON(&action)

		if err != nil {
			log.Println("[V1] Error reading JSON.")
			writeError(conn, 2001)
			break
		}

		if action.Type == "" {
			writeError(conn, 3001)
			continue
		}

		if action.Payload == nil {
			writeError(conn, 3001)
			continue
		}

		switch action.Type {
		case authorizationSignInRequest:
			handleAuthorizationSignInRequest(conn, &action)
		case messagingBroadcastRequest:
			handleMessagingBroadcastRequest(conn, &action)
		case verificationRequestCodeRequest:
			handleVerificationRequestCodeRequest(conn, &action)
		case verificationSubmitCodeRequest:
			handleVerificationSubmitCodeRequest(conn, &action)
		default:
			writeError(conn, 3002)
		}
	}
}

func handleAuthorizationSignInRequest(conn *websocket.Conn, action *Action) {
	countryCode := action.Payload["country_code"].(string)
	phoneNumber := action.Payload["phone_number"].(string)
	verificationCode := action.Payload["code"].(string)
	client, ok := findVerifiedClient(countryCode, phoneNumber, verificationCode)

	if !ok {
		log.Println("[V1] Couldn't find such client.")
		writeFailure(conn, authorizationSignInFailure, []string{"couldn't find you"})
		return
	}

	client.IsSignedIn = true
	clients[conn] = client

	for pendingClient := range pendingActionQueue {
		if pendingClient.CountryCode == client.CountryCode &&
			pendingClient.PhoneNumber == client.PhoneNumber {

			for _, pendingAction := range pendingActionQueue[pendingClient] {
				conn.WriteJSON(*pendingAction)
			}

			delete(pendingActionQueue, pendingClient)
		}
	}

	writeEmptyAction(conn, authorizationSignInSuccess)
}

func handleMessagingBroadcastRequest(conn *websocket.Conn, action *Action) {
	if _, ok := clients[conn]; !ok {
		writeEmptyAction(conn, messagingBroadcastFailure)
		return
	}

	if !clients[conn].IsSignedIn {
		writeEmptyAction(conn, messagingBroadcastFailure)
		return
	}

	sender := make(map[string]string)
	sender["country_code"] = clients[conn].CountryCode
	sender["phone_number"] = clients[conn].PhoneNumber

	payload := make(map[string]interface{})
	payload["message"] = action.Payload["message"]
	payload["sender"] = sender

	response := &Action{
		Payload: payload,
		Type:    messagingBroadcastSuccess,
	}

	broadcastChan <- response
}

func handleVerificationRequestCodeRequest(conn *websocket.Conn, action *Action) {
	countryCode, ok := action.Payload["country_code"].(string)

	if !ok {
		log.Println("country_code is required in Payload.")
		writeEmptyAction(conn, verificationRequestCodeFailure)
		return
	}

	phoneNumber, ok := action.Payload["phone_number"].(string)

	if !ok {
		log.Println("phone_number is required in Payload.")
		writeEmptyAction(conn, verificationRequestCodeFailure)
		return
	}

	resp, err := postTwilioVerificationStart(countryCode, phoneNumber)

	if err != nil {
		writeEmptyAction(conn, verificationRequestCodeFailure)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		log.Println("[V1] Failed to read Twilio API response body.")
		writeEmptyAction(conn, verificationRequestCodeFailure)
		return
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)

	if err != nil {
		log.Println("[V1] Failed to parse Twilio verification start response body JSON.")
		writeEmptyAction(conn, verificationRequestCodeFailure)
		return
	}

	if !r["success"].(bool) {
		log.Println("[V1] Twilio verification start API call failed.")
		writeEmptyAction(conn, verificationRequestCodeFailure)
		return
	}

	if !findClient(countryCode, phoneNumber) {
		if !createClient(countryCode, phoneNumber) {
			log.Println("[V1] Failed to create client.")
			writeEmptyAction(conn, verificationRequestCodeFailure)
			return
		}
	}

	client := &Client{
		CountryCode:      countryCode,
		PhoneNumber:      phoneNumber,
		VerificationCode: "",
	}
	updateClientVerificationCode(client, client.VerificationCode)
	clients[conn] = client
	writeEmptyAction(conn, verificationRequestCodeSuccess)
}

func handleVerificationSubmitCodeRequest(conn *websocket.Conn, action *Action) {
	code, ok := action.Payload["code"].(string)

	if !ok {
		writeEmptyAction(conn, verificationSubmitCodeFailure)
		return
	}

	client, ok := clients[conn]

	if !ok {
		writeEmptyAction(conn, verificationSubmitCodeFailure)
		return
	}

	if client.CountryCode == "" {
		writeEmptyAction(conn, verificationSubmitCodeFailure)
		return
	}

	if client.PhoneNumber == "" {
		writeEmptyAction(conn, verificationSubmitCodeFailure)
		return
	}

	if client.VerificationCode == "" {
		resp := getTwilioVerificationCheck(client.CountryCode, client.PhoneNumber, code)

		if resp == nil {
			writeEmptyAction(conn, verificationSubmitCodeFailure)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			log.Println("[V1] Failed to read Twilio verification check API response.")
			writeEmptyAction(conn, verificationSubmitCodeFailure)
			return
		}

		r := map[string]interface{}{}
		err = json.Unmarshal(body, &r)

		if err != nil {
			log.Println("[V1] Failed to parse Twilio verification check response body JSON.")
			writeEmptyAction(conn, verificationSubmitCodeFailure)
			return
		}

		if !r["success"].(bool) {
			log.Println("[V1] Twilio verification check failed.")
			writeEmptyAction(conn, verificationSubmitCodeFailure)
			return
		}

		if !updateClientVerificationCode(client, code) {
			log.Println("[V1] Failed to update verification code.")
			writeEmptyAction(conn, verificationSubmitCodeFailure)
			return
		}

		client.VerificationCode = code
	} else {
		if client.VerificationCode != code {
			log.Println("[V1] Verification code doesn't match.")
			writeEmptyAction(conn, verificationSubmitCodeFailure)
			return
		}
	}

	writeEmptyAction(conn, verificationSubmitCodeSuccess)
}
