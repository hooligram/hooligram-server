package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func getTwilioVerificationCheck(countryCode, phoneNumber, verificationCode string) *http.Response {
	url := "https://api.authy.com/protected/json/phones/verification/check"
	url += "?country_code=" + countryCode
	url += "&phone_number=" + phoneNumber
	url += "&verification_code=" + verificationCode

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Println("[API] Failed to make Twilio verification check request.")
		return nil
	}

	req.Header.Add("X-Authy-API-Key", twilioAPIKey)
	resp, err := httpClient.Do(req)

	if err != nil {
		log.Println("[API] Failed to read Twilio verification check API response.")
		return nil
	}

	return resp
}

func postTwilioVerificationStart(countryCode, phoneNumber string) *http.Response {
	b, err := json.Marshal(map[string]interface{}{
		"api_key":      twilioAPIKey,
		"country_code": countryCode,
		"phone_number": phoneNumber,
		"via":          "sms",
	})

	if err != nil {
		log.Println("[API] Failed to encode Twilio JSON request payload.")
		return nil
	}

	resp, err := http.Post(
		"https://api.authy.com/protected/json/phones/verification/start",
		"application/json",
		bytes.NewReader(b),
	)

	if err != nil {
		log.Println("[API] Failed to start Twilio verification API call.")
		return nil
	}

	return resp
}
