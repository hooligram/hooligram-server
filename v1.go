package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
)

func v1(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Can't upgrade to WebSocket connection.")
		return
	}

	clients[conn] = &Client{}
	defer conn.Close()

	for {
		action := Action{}
		err = conn.ReadJSON(&action)

		if err != nil {
			log.Println("Error reading JSON.")
			writeError(conn, 2001)
			continue
		}

		if action.Type == "" {
			log.Println("Action Type is required.")
			writeError(conn, 3001)
			continue
		}

		if action.Payload == nil {
			log.Println("Action Payload is required.")
			writeError(conn, 3001)
			continue
		}

		switch action.Type {
		case verificationRequestCodeRequest:
			countryCode, ok := action.Payload["country_code"].(string)

			if !ok {
				log.Println("country_code is required in Payload.")
				writeEmptyAction(conn, verificationRequestCodeFailure)
				break
			}

			phoneNumber, ok := action.Payload["phone_number"].(string)

			if !ok {
				log.Println("phone_number is required in Payload.")
				writeEmptyAction(conn, verificationRequestCodeFailure)
				break
			}

			b, err := json.Marshal(map[string]interface{}{
				"api_key":      twilioAPIKey,
				"country_code": countryCode,
				"phone_number": phoneNumber,
				"via":          "sms",
			})

			if err != nil {
				log.Println(err)
				writeEmptyAction(conn, verificationRequestCodeFailure)
				break
			}

			resp, err := http.Post(
				"https://api.authy.com/protected/json/phones/verification/start",
				"application/json",
				bytes.NewReader(b),
			)

			if err != nil {
				log.Println(err)
				writeEmptyAction(conn, verificationRequestCodeFailure)
				break
			}

			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			if err != nil {
				log.Println(err)
				writeEmptyAction(conn, verificationRequestCodeFailure)
				break
			}

			r := map[string]interface{}{}
			json.Unmarshal(body, &r)

			if !r["success"].(bool) {
				log.Println("Failed to send verification code.")
				writeEmptyAction(conn, verificationRequestCodeFailure)
				break
			}

			clients[conn] = &Client{CountryCode: countryCode, PhoneNumber: phoneNumber}
			conn.WriteJSON(Action{
				map[string]interface{}{},
				verificationRequestCodeSuccess,
			})
		case verificationSubmitCodeRequest:
			code, ok := action.Payload["code"].(string)

			if !ok {
				log.Println("code is required in Payload.")
				writeEmptyAction(conn, verificationSubmitCodeFailure)
				break
			}

			client, ok := clients[conn]

			if !ok {
				log.Println("Unknown client.")
				writeEmptyAction(conn, verificationSubmitCodeFailure)
				break
			}

			if client.CountryCode == "" {
				log.Println("Client's country code record is missing.")
				writeEmptyAction(conn, verificationSubmitCodeFailure)
				break
			}

			if client.PhoneNumber == "" {
				log.Println("Client's phone number record is missing.")
				writeEmptyAction(conn, verificationSubmitCodeFailure)
				break
			}

			if client.VerificationCode == "" {
				url := "https://api.authy.com/protected/json/phones/verification/check"
				url += "?country_code=" + client.CountryCode
				url += "&phone_number=" + client.PhoneNumber
				url += "&verification_code=" + code

				req, err := http.NewRequest("GET", url, nil)

				if err != nil {
					writeEmptyAction(conn, verificationSubmitCodeFailure)
					break
				}

				req.Header.Add("X-Authy-API-Key", twilioAPIKey)
				resp, err := httpClient.Do(req)

				if err != nil {
					log.Println("Failed to send verify code.")
					writeEmptyAction(conn, verificationSubmitCodeFailure)
					break
				}

				body, err := ioutil.ReadAll(resp.Body)
				resp.Body.Close()

				r := map[string]interface{}{}
				json.Unmarshal(body, &r)

				if !r["success"].(bool) {
					log.Println("Verification code is incorrect.")
					writeEmptyAction(conn, verificationRequestCodeFailure)
					break
				}

				client.VerificationCode = code
			} else {
				if client.VerificationCode != code {
					log.Println("Verification code doesn't match the record.")
					writeEmptyAction(conn, verificationRequestCodeFailure)
					break
				}
			}

			conn.WriteJSON(Action{
				map[string]interface{}{},
				verificationSubmitCodeSuccess,
			})
		default:
			log.Println("Not supported Action Type.")
			writeError(conn, 3002)
		}
	}
}
