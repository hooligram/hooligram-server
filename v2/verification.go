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
	actionID := action.ID
	if actionID == "" {
		return verificationRequestCodeFailure(client, actionID, "id not in action")
	}

	countryCode, ok := action.Payload["country_code"].(string)
	if !ok {
		return verificationRequestCodeFailure(client, actionID, "country_code not in payload")
	}

	phoneNumber, ok := action.Payload["phone_number"].(string)
	if !ok {
		return verificationRequestCodeFailure(client, actionID, "phone_number not in payload")
	}

	ok, err := client.Register(countryCode, phoneNumber)
	if err != nil {
		utils.LogBody(v2Tag, "error registering client. "+err.Error())
		return verificationRequestCodeFailure(client, actionID, "server error")
	}

	if !ok {
		return verificationRequestCodeFailure(client, actionID, "unable to register")
	}

	resp, err := api.PostTwilioVerificationStart(countryCode, phoneNumber)
	if err != nil {
		utils.LogBody(v2Tag, "error posting twilio verification start. "+err.Error())
		return verificationRequestCodeFailure(client, actionID, "server error")
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.LogBody(v2Tag, "error reading response body. "+err.Error())
		return verificationRequestCodeFailure(client, actionID, "server error")
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		utils.LogBody(v2Tag, "error parsing json. "+err.Error())
		return verificationRequestCodeFailure(client, actionID, "server error")
	}

	if !r["success"].(bool) {
		return verificationRequestCodeFailure(client, actionID, "server error")
	}

	success := actions.VerificationRequestCodeSuccess(actionID)
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
	actionID := action.ID
	if actionID == "" {
		return verificationSubmitCodeFailure(client, actionID, "id not in action")
	}

	verificationCode, ok := action.Payload["verification_code"].(string)
	if !ok {
		return verificationSubmitCodeFailure(client, actionID, "verification_code not in payload")
	}

	isVerified, err := client.IsVerified()
	if err != nil {
		utils.LogBody(v2Tag, "error checking if client is verified. "+err.Error())
		return verificationSubmitCodeFailure(client, actionID, "server error")
	}

	if isVerified {
		storedVerificationCode, err := client.GetVerificationCode()
		if err != nil {
			utils.LogBody(v2Tag, "error getting client verification code. "+err.Error())
			return verificationSubmitCodeFailure(client, actionID, "server error")
		}

		if verificationCode == storedVerificationCode {
			success := actions.VerificationSubmitCodeSuccess(actionID)
			client.WriteJSON(success)
			return success
		}
		{
			return verificationSubmitCodeFailure(client, actionID, "incorrect verification code")
		}
	}

	clientRow, err := db.ReadClientByID(client.GetID())
	if err != nil {
		utils.LogBody(v2Tag, "error reading client by id. "+err.Error())
		return verificationSubmitCodeFailure(client, actionID, "server error")
	}

	resp, err := api.GetTwilioVerificationCheck(
		clientRow.CountryCode,
		clientRow.PhoneNumber,
		verificationCode,
	)
	if err != nil {
		utils.LogBody(v2Tag, "twilio verification check error. "+err.Error())
		return verificationSubmitCodeFailure(client, actionID, "server error")
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.LogBody(v2Tag, "error reading response body. "+err.Error())
		return verificationSubmitCodeFailure(client, actionID, "server error")
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		utils.LogBody(v2Tag, "error parsing json. "+err.Error())
		return verificationSubmitCodeFailure(client, actionID, "server error")
	}

	if !r["success"].(bool) {
		return verificationSubmitCodeFailure(client, actionID, "incorrect verification code")
	}

	err = db.UpdateClientVerificationCode(client.GetID(), verificationCode)
	if err != nil {
		utils.LogBody(v2Tag, "error setting client as verified. "+err.Error())
		return verificationSubmitCodeFailure(client, actionID, "server error")
	}

	success := actions.VerificationSubmitCodeSuccess(actionID)
	client.WriteJSON(success)
	return success
}

////////////
// HELPER //
////////////

func verificationRequestCodeFailure(client *clients.Client, actionID, err string) *actions.Action {
	failure := actions.VerificationRequestCodeFailure(actionID, []string{err})
	client.WriteJSON(failure)
	return failure
}

func verificationSubmitCodeFailure(client *clients.Client, actionID, err string) *actions.Action {
	failure := actions.VerificationSubmitCodeFailure(actionID, []string{err})
	client.WriteJSON(failure)
	return failure
}
