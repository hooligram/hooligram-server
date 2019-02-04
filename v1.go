package main

import (
	"bytes"
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
			continue
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

	if !findVerifiedClient(countryCode, phoneNumber, verificationCode) {
		writeEmptyAction(conn, authorizationSignInFailure)
		return
	}

	client := &Client{
		CountryCode:      countryCode,
		IsSignedIn:       true,
		PhoneNumber:      phoneNumber,
		VerificationCode: verificationCode,
	}
	clients[conn] = client

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

	b, err := json.Marshal(map[string]interface{}{
		"api_key":      twilioAPIKey,
		"country_code": countryCode,
		"phone_number": phoneNumber,
		"via":          "sms",
	})

	if err != nil {
		log.Println("[V1] Failed to encode Twilio JSON request payload.")
		writeEmptyAction(conn, verificationRequestCodeFailure)
		return
	}

	resp, err := http.Post(
		"https://api.authy.com/protected/json/phones/verification/start",
		"application/json",
		bytes.NewReader(b),
	)

	if err != nil {
		log.Println("[V1] Failed to start Twilio verification API call.")
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
	json.Unmarshal(body, &r)

	if !r["success"].(bool) {
		writeEmptyAction(conn, verificationRequestCodeFailure)
		return
	}

	if !findClient(countryCode, phoneNumber) {
		if !createClient(countryCode, phoneNumber) {
			writeEmptyAction(conn, verificationRequestCodeFailure)
			return
		}
	}

	client := &Client{
		CountryCode:      countryCode,
		PhoneNumber:      phoneNumber,
		VerificationCode: "",
	}
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
		url := "https://api.authy.com/protected/json/phones/verification/check"
		url += "?country_code=" + client.CountryCode
		url += "&phone_number=" + client.PhoneNumber
		url += "&verification_code=" + code

		req, err := http.NewRequest("GET", url, nil)

		if err != nil {
			log.Println("[V1] Failed to make Twilio verification check request.")
			writeEmptyAction(conn, verificationSubmitCodeFailure)
			return
		}

		req.Header.Add("X-Authy-API-Key", twilioAPIKey)
		resp, err := httpClient.Do(req)

		if err != nil {
			log.Println("[V1] Failed to send Twilio verification check API call.")
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
		json.Unmarshal(body, &r)

		if !r["success"].(bool) {
			writeEmptyAction(conn, verificationRequestCodeFailure)
			return
		}
	} else {
		if client.VerificationCode != code {
			writeEmptyAction(conn, verificationRequestCodeFailure)
			return
		}
	}

	if !updateClientVerificationCode(client, code) {
		writeEmptyAction(conn, verificationRequestCodeFailure)
		return
	}

	client.VerificationCode = code
	writeEmptyAction(conn, verificationSubmitCodeSuccess)
}
