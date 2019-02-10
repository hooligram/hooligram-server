package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func getTwilioVerificationCheck(countryCode, phoneNumber, verificationCode string) (*http.Response, error) {
	url := "https://api.authy.com/protected/json/phones/verification/check"
	url += "?country_code=" + countryCode
	url += "&phone_number=" + phoneNumber
	url += "&verification_code=" + verificationCode

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Authy-API-Key", twilioAPIKey)
	resp, err := httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func postTwilioVerificationStart(countryCode, phoneNumber string) (*http.Response, error) {
	url := "https://api.authy.com/protected/json/phones/verification/start"
	b, err := json.Marshal(map[string]interface{}{
		"country_code": countryCode,
		"phone_number": phoneNumber,
		"via":          "sms",
	})

	if err != nil {
		log.Println("[API] Failed to construct Twilio verification start JSON body.")
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))

	if err != nil {
		log.Println("[API] Failed to create Twilio verification start POST request.")
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Authy-API-Key", twilioAPIKey)
	resp, err := httpClient.Do(req)

	if err != nil {
		log.Println("[API] Failed to do Twilio verification start API call.")
		return nil, err
	}

	return resp, nil
}
