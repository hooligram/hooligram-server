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

////////////////////////////////////////////////
// HANDLER: VERIFICATION_REQUEST_CODE_REQUEST //
////////////////////////////////////////////////

func handleVerificationRequestCodeRequest(
	client *clients.Client,
	action *actions.Action,
) *actions.Action {
	countryCode, ok := action.Payload["country_code"].(string)
	if !ok {
		return verificationRequestCodeFailure(client, "country_code not in payload")
	}

	phoneNumber, ok := action.Payload["phone_number"].(string)
	if !ok {
		return verificationRequestCodeFailure(client, "phone_number not in payload")
	}

	ok, err := client.Register(countryCode, phoneNumber)
	if err != nil {
		utils.LogBody(v2Tag, "error registering client. "+err.Error())
		return verificationRequestCodeFailure(client, "server error")
	}

	if !ok {
		return verificationRequestCodeFailure(client, "unable to register")
	}

	resp, err := api.PostTwilioVerificationStart(countryCode, phoneNumber)
	if err != nil {
		utils.LogBody(v2Tag, "error posting twilio verification start. "+err.Error())
		return verificationRequestCodeFailure(client, "server error")
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.LogBody(v2Tag, "error reading response body. "+err.Error())
		return verificationRequestCodeFailure(client, "server error")
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		utils.LogBody(v2Tag, "error parsing json. "+err.Error())
		return verificationRequestCodeFailure(client, "server error")
	}

	if !r["success"].(bool) {
		return verificationRequestCodeFailure(client, "server error")
	}

	success := actions.VerificationRequestCodeSuccess()
	client.WriteJSON(success)
	return success
}

///////////////////////////////////////////////
// HANDLER: VERIFICATION_SUBMIT_CODE_REQUEST //
///////////////////////////////////////////////

func handleVerificationSubmitCodeRequest(
	client *clients.Client,
	action *actions.Action,
) *actions.Action {
	verificationCode, ok := action.Payload["verification_code"].(string)
	if !ok {
		return verificationSubmitCodeFailure(client, "verification_code not in payload")
	}

	isVerified, err := client.IsVerified()
	if err != nil {
		utils.LogBody(v2Tag, "error checking if client is verified. "+err.Error())
		return verificationSubmitCodeFailure(client, "server error")
	}

	if isVerified {
		storedVerificationCode, err := client.GetVerificationCode()
		if err != nil {
			utils.LogBody(v2Tag, "error getting client verification code. "+err.Error())
			return verificationSubmitCodeFailure(client, "server error")
		}

		if verificationCode == storedVerificationCode {
			success := actions.VerificationSubmitCodeSuccess()
			client.WriteJSON(success)
			return success
		}
		{
			return verificationSubmitCodeFailure(client, "incorrect verification code")
		}
	}

	clientRow, err := db.ReadClientByID(client.GetID())
	if err != nil {
		utils.LogBody(v2Tag, "error reading client by id. "+err.Error())
		return verificationSubmitCodeFailure(client, "server error")
	}

	resp, err := api.GetTwilioVerificationCheck(
		clientRow.CountryCode,
		clientRow.PhoneNumber,
		verificationCode,
	)
	if err != nil {
		utils.LogBody(v2Tag, "twilio verification check error. "+err.Error())
		return verificationSubmitCodeFailure(client, "server error")
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.LogBody(v2Tag, "error reading response body. "+err.Error())
		return verificationSubmitCodeFailure(client, "server error")
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		utils.LogBody(v2Tag, "error parsing json. "+err.Error())
		return verificationSubmitCodeFailure(client, "server error")
	}

	if !r["success"].(bool) {
		return verificationSubmitCodeFailure(client, "incorrect verification code")
	}

	err = db.UpdateClientVerificationCode(client.GetID(), verificationCode)
	if err != nil {
		utils.LogBody(v2Tag, "error setting client as verified. "+err.Error())
		return verificationSubmitCodeFailure(client, "server error")
	}

	success := actions.VerificationSubmitCodeSuccess()
	client.WriteJSON(success)
	return success
}

////////////
// HELPER //
////////////

func verificationRequestCodeFailure(client *clients.Client, err string) *actions.Action {
	failure := actions.VerificationRequestCodeFailure([]string{err})
	client.WriteJSON(failure)
	return failure
}

func verificationSubmitCodeFailure(client *clients.Client, err string) *actions.Action {
	failure := actions.VerificationSubmitCodeFailure([]string{err})
	client.WriteJSON(failure)
	return failure
}
