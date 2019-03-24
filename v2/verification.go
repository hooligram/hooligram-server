package v2

import (
	"encoding/json"
	"io/ioutil"

	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/api"
	"github.com/hooligram/hooligram-server/clients"
	"github.com/hooligram/hooligram-server/db"
	"github.com/hooligram/hooligram-server/utils"
)

func handleVerificationRequestCodeRequest(
	client *clients.Client,
	action *actions.Action,
) *actions.Action {
	countryCode, ok := action.Payload["country_code"].(string)
	if !ok {
		failure := actions.CreateVerificationRequestCodeFailure(
			[]string{"country_code not in payload"},
		)
		client.WriteJSON(failure)
		return failure
	}

	phoneNumber, ok := action.Payload["phone_number"].(string)
	if !ok {
		failure := actions.CreateVerificationRequestCodeFailure(
			[]string{"phone_number not in payload"},
		)
		client.WriteJSON(failure)
		return failure
	}

	ok, err := client.Register(countryCode, phoneNumber)
	if err != nil {
		utils.LogBody(v2Tag, "error registering client. "+err.Error())
		failure := actions.CreateVerificationRequestCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	if !ok {
		failure := actions.CreateVerificationRequestCodeFailure([]string{"unable to register"})
		client.WriteJSON(failure)
		return failure
	}

	resp, err := api.PostTwilioVerificationStart(countryCode, phoneNumber)
	if err != nil {
		utils.LogBody(v2Tag, "error posting twilio verification start. "+err.Error())
		failure := actions.CreateVerificationRequestCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.LogBody(v2Tag, "error reading response body. "+err.Error())
		failure := actions.CreateVerificationRequestCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		utils.LogBody(v2Tag, "error parsing json. "+err.Error())
		failure := actions.CreateVerificationRequestCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	if !r["success"].(bool) {
		failure := actions.CreateVerificationRequestCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	success := actions.CreateVerificationRequestCodeSuccess()
	client.WriteJSON(success)
	return success
}

func handleVerificationSubmitCodeRequest(
	client *clients.Client,
	action *actions.Action,
) *actions.Action {
	code, ok := action.Payload["code"].(string)
	if !ok {
		failure := actions.CreateVerificationSubmitCodeFailure([]string{"code not in payload"})
		client.WriteJSON(failure)
		return failure
	}

	isVerified, err := client.IsVerified()
	if err != nil {
		utils.LogBody(v2Tag, "error checking if client is verified. "+err.Error())
		failure := actions.CreateVerificationSubmitCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	if isVerified {
		verificationCode, err := client.GetVerificationCode()
		if err != nil {
			utils.LogBody(v2Tag, "error getting client verification code. "+err.Error())
			failure := actions.CreateVerificationSubmitCodeFailure([]string{"server error"})
			client.WriteJSON(failure)
			return failure
		}

		if code == verificationCode {
			success := actions.CreateVerificationSubmitCodeSuccess()
			client.WriteJSON(success)
			return success
		}
		{
			failure := actions.CreateVerificationSubmitCodeFailure([]string{"incorrect verification code"})
			client.WriteJSON(failure)
			return failure
		}
	}

	clientRow, err := db.ReadClientByID(client.GetID())
	if err != nil {
		utils.LogBody(v2Tag, "error reading client by id. "+err.Error())
		failure := actions.CreateVerificationSubmitCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	resp, err := api.GetTwilioVerificationCheck(clientRow.CountryCode, clientRow.PhoneNumber, code)
	if err != nil {
		utils.LogBody(v2Tag, "twilio verification check error. "+err.Error())
		failure := actions.CreateVerificationSubmitCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.LogBody(v2Tag, "error reading response body. "+err.Error())
		failure := actions.CreateVerificationSubmitCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		utils.LogBody(v2Tag, "error parsing json. "+err.Error())
		failure := actions.CreateVerificationSubmitCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	if !r["success"].(bool) {
		failure := actions.CreateVerificationSubmitCodeFailure(
			[]string{"incorrect verification code"},
		)
		client.WriteJSON(failure)
		return failure
	}

	err = db.UpdateClientVerificationCode(client.GetID(), code)
	if err != nil {
		utils.LogBody(v2Tag, "error setting client as verified. "+err.Error())
		failure := actions.CreateVerificationSubmitCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	success := actions.CreateVerificationSubmitCodeSuccess()
	client.WriteJSON(success)
	return success
}
