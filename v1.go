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
		case "VERIFICATION_VERIFY_PHONE_NO":
			countryCode, ok := action.Payload["country_code"].(float64)

			if !ok {
				log.Println("country_code is required in Payload.")
				writeError(conn, 3001)
				break
			}

			phoneNumber, ok := action.Payload["phone_number"].(float64)

			if !ok {
				log.Println("phone_number is required in Payload.")
				writeError(conn, 3001)
				break
			}

			b, err := json.Marshal(map[string]interface{}{
				"api_key":      twilioAPIKey,
				"country_code": int(countryCode),
				"phone_number": int(phoneNumber),
				"via":          "sms",
			})

			if err != nil {
				log.Println(err)
				writeError(conn, 5000)
				break
			}

			resp, err := http.Post(
				"https://api.authy.com/protected/json/phones/verification/start",
				"application/json",
				bytes.NewReader(b),
			)

			if err != nil {
				log.Println(err)
				writeError(conn, 5000)
				break
			}

			body, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()

			if err != nil {
				log.Println(err)
				writeError(conn, 5000)
				break
			}

			r := map[string]interface{}{}
			json.Unmarshal(body, &r)

			if !r["success"].(bool) {
				log.Println("Failed to send verification code.")
				conn.WriteJSON(Action{
					map[string]interface{}{},
					"VERIFICATION_VERIFICATION_CODE_SENT_FAILED",
				})
				break
			}

			conn.WriteJSON(Action{
				map[string]interface{}{},
				"VERIFICATION_VERIFICATION_CODE_SENT",
			})
		default:
			log.Println("Not supported Action Type.")
			writeError(conn, 3002)
		}
	}
}
