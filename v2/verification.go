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
		failure := actions.VerificationRequestCodeFailure(
			[]string{"country_code not in payload"},
		)
		client.WriteJSON(failure)
		return failure
	}

	phoneNumber, ok := action.Payload["phone_number"].(string)
	if !ok {
		failure := actions.VerificationRequestCodeFailure(
			[]string{"phone_number not in payload"},
		)
		client.WriteJSON(failure)
		return failure
	}

	ok, err := client.Register(countryCode, phoneNumber)
	if err != nil {
		utils.LogBody(v2Tag, "error registering client. "+err.Error())
		failure := actions.VerificationRequestCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	if !ok {
		failure := actions.VerificationRequestCodeFailure([]string{"unable to register"})
		client.WriteJSON(failure)
		return failure
	}

	resp, err := api.PostTwilioVerificationStart(countryCode, phoneNumber)
	if err != nil {
		utils.LogBody(v2Tag, "error posting twilio verification start. "+err.Error())
		failure := actions.VerificationRequestCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.LogBody(v2Tag, "error reading response body. "+err.Error())
		failure := actions.VerificationRequestCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		utils.LogBody(v2Tag, "error parsing json. "+err.Error())
		failure := actions.VerificationRequestCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	if !r["success"].(bool) {
		failure := actions.VerificationRequestCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	success := actions.VerificationRequestCodeSuccess()
	client.WriteJSON(success)
	return success
}

func handleVerificationSubmitCodeRequest(
	client *clients.Client,
	action *actions.Action,
) *actions.Action {
	verificationCode, ok := action.Payload["verification_code"].(string)
	if !ok {
		failure := actions.VerificationSubmitCodeFailure(
			[]string{"verification_code not in payload"},
		)
		client.WriteJSON(failure)
		return failure
	}

	isVerified, err := client.IsVerified()
	if err != nil {
		utils.LogBody(v2Tag, "error checking if client is verified. "+err.Error())
		failure := actions.VerificationSubmitCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	if isVerified {
		storedVerificationCode, err := client.GetVerificationCode()
		if err != nil {
			utils.LogBody(v2Tag, "error getting client verification code. "+err.Error())
			failure := actions.VerificationSubmitCodeFailure([]string{"server error"})
			client.WriteJSON(failure)
			return failure
		}

		if verificationCode == storedVerificationCode {
			success := actions.VerificationSubmitCodeSuccess()
			client.WriteJSON(success)
			return success
		}
		{
			failure := actions.VerificationSubmitCodeFailure(
				[]string{"incorrect verification code"},
			)
			client.WriteJSON(failure)
			return failure
		}
	}

	clientRow, err := db.ReadClientByID(client.GetID())
	if err != nil {
		utils.LogBody(v2Tag, "error reading client by id. "+err.Error())
		failure := actions.VerificationSubmitCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	resp, err := api.GetTwilioVerificationCheck(
		clientRow.CountryCode,
		clientRow.PhoneNumber,
		verificationCode,
	)
	if err != nil {
		utils.LogBody(v2Tag, "twilio verification check error. "+err.Error())
		failure := actions.VerificationSubmitCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.LogBody(v2Tag, "error reading response body. "+err.Error())
		failure := actions.VerificationSubmitCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		utils.LogBody(v2Tag, "error parsing json. "+err.Error())
		failure := actions.VerificationSubmitCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	if !r["success"].(bool) {
		failure := actions.VerificationSubmitCodeFailure(
			[]string{"incorrect verification code"},
		)
		client.WriteJSON(failure)
		return failure
	}

	err = db.UpdateClientVerificationCode(client.GetID(), verificationCode)
	if err != nil {
		utils.LogBody(v2Tag, "error setting client as verified. "+err.Error())
		failure := actions.VerificationSubmitCodeFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	success := actions.VerificationSubmitCodeSuccess()
	client.WriteJSON(success)
	return success
}
