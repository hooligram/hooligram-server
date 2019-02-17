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
	client, err := signIn(conn, countryCode, phoneNumber, verificationCode)

	if err != nil {
		log.Println("[V1] Couldn't sign in client.")
		log.Println("[V1]", err.Error())
		writeFailure(conn, authorizationSignInFailure, []string{"couldn't sign in you"})
		return
	}

	writeQueuedActions(client)
	client.writeEmptyAction(authorizationSignInSuccess)
}

func handleMessagingBroadcastRequest(conn *websocket.Conn, action *Action) {
	client, err := getSignedInClient(conn)

	if err != nil {
		client.writeFailure(messagingBroadcastFailure, []string{err.Error()})
		return
	}

	message, ok := action.Payload["message"].(string)

	if !ok {
		client.writeFailure(messagingBroadcastFailure, []string{"you forgot your message"})
		return
	}

	broadcastChan <- constructBroadcastAction(client, message)
}

func handleVerificationRequestCodeRequest(conn *websocket.Conn, action *Action) {
	errors := []string{}
	countryCode, ok := action.Payload["country_code"].(string)

	if !ok {
		errors = append(errors, "you need to include country_code in payload")
	}

	phoneNumber, ok := action.Payload["phone_number"].(string)

	if !ok {
		errors = append(errors, "you need to include phone_number in payload")
	}

	if len(errors) > 0 {
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	resp, err := postTwilioVerificationStart(countryCode, phoneNumber)

	if err != nil {
		errors = append(errors, err.Error())
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		errors = append(errors, err.Error())
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)

	if err != nil {
		errors = append(errors, err.Error())
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	if !r["success"].(bool) {
		errors = append(errors, "i failed to make verification start api call")
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	client, err := getOrCreateClient(countryCode, phoneNumber)

	if err != nil {
		errors = append(errors, err.Error())
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	err = unverifyClient(client, conn)

	if err != nil {
		errors = append(errors, err.Error())
		writeFailure(conn, verificationRequestCodeFailure, errors)
		return
	}

	writeEmptyAction(conn, verificationRequestCodeSuccess)
}

func handleVerificationSubmitCodeRequest(conn *websocket.Conn, action *Action) {
	var errors []string
	code, ok := action.Payload["code"].(string)

	if !ok {
		errors = append(errors, "you need to include code in payload")
		writeFailure(conn, verificationSubmitCodeFailure, errors)
		return
	}

	client, err := getClient(conn)

	if err != nil {
		errors = append(errors, err.Error())
		writeFailure(conn, verificationSubmitCodeFailure, errors)
		return
	}

	if !client.isVerified() {
		resp, err := getTwilioVerificationCheck(client.CountryCode, client.PhoneNumber, code)

		if err != nil {
			errors = append(errors, err.Error())
			client.writeFailure(verificationSubmitCodeFailure, errors)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			errors = append(errors, err.Error())
			client.writeFailure(verificationSubmitCodeFailure, errors)
			return
		}

		r := map[string]interface{}{}
		err = json.Unmarshal(body, &r)

		if err != nil {
			errors = append(errors, err.Error())
			client.writeFailure(verificationSubmitCodeFailure, errors)
			return
		}

		if !r["success"].(bool) {
			unverifyClient(client, conn)
			errors = append(errors, "twilio verification check failed")
			client.writeFailure(verificationSubmitCodeFailure, errors)
			return
		}
	} else {
		if client.VerificationCode != code {
			unverifyClient(client, conn)
			errors = append(errors, "verification code doesn't match")
			client.writeFailure(verificationSubmitCodeFailure, errors)
			return
		}
	}

	verifyClient(client, conn, code)
	writeEmptyAction(conn, verificationSubmitCodeSuccess)
}
