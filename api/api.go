package api

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/hooligram/hooligram-server/globals"
)

const apiTag = "api"

var httpClient = &http.Client{}

// GetTwilioVerificationCheck .
func GetTwilioVerificationCheck(
	countryCode,
	phoneNumber,
	verificationCode string,
) (*http.Response, error) {
	url := "https://api.authy.com/protected/json/phones/verification/check"
	url += "?country_code=" + countryCode
	url += "&phone_number=" + phoneNumber
	url += "&verification_code=" + verificationCode

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Authy-API-Key", globals.TwilioAPIKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// PostTwilioVerificationStart .
func PostTwilioVerificationStart(countryCode, phoneNumber string) (*http.Response, error) {
	url := "https://api.authy.com/protected/json/phones/verification/start"
	b, err := json.Marshal(map[string]interface{}{
		"country_code": countryCode,
		"phone_number": phoneNumber,
		"via":          "sms",
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Authy-API-Key", globals.TwilioAPIKey)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
