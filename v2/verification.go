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

func handleVerificationRequestCodeRequest(client *clients.Client, action *actions.Action) {
	countryCode, ok := action.Payload["country_code"].(string)
	if !ok {
		client.WriteFailure(
			actions.VerificationRequestCodeFailure,
			[]string{"country_code not in payload"},
		)
		return
	}

	phoneNumber, ok := action.Payload["phone_number"].(string)
	if !ok {
		client.WriteFailure(
			actions.VerificationRequestCodeFailure,
			[]string{"phone_number not in payload"},
		)
		return
	}

	ok, err := client.Register(countryCode, phoneNumber)
	if err != nil {
		utils.LogBody(v2Tag, "register client error. "+err.Error())
		client.WriteFailure(
			actions.VerificationRequestCodeFailure,
			[]string{"server error"},
		)
		return
	}

	if !ok {
		client.WriteFailure(
			actions.VerificationRequestCodeFailure,
			[]string{"unable to register"},
		)
		return
	}

	resp, err := api.PostTwilioVerificationStart(countryCode, phoneNumber)
	if err != nil {
		utils.LogBody(v2Tag, "twilio verification start error. "+err.Error())
		client.WriteFailure(
			actions.VerificationRequestCodeFailure,
			[]string{"server error"},
		)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.LogBody(v2Tag, "error reading response body. "+err.Error())
		client.WriteFailure(
			actions.VerificationRequestCodeFailure,
			[]string{"server error"},
		)
		return
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		utils.LogBody(v2Tag, "error parsing json. "+err.Error())
		client.WriteFailure(
			actions.VerificationRequestCodeFailure,
			[]string{"server error"},
		)
		return
	}

	if !r["success"].(bool) {
		utils.LogBody(v2Tag, "twilio responded with failure. "+err.Error())
		client.WriteFailure(
			actions.VerificationRequestCodeFailure,
			[]string{"server error"},
		)
		return
	}

	client.WriteEmptyAction(actions.VerificationRequestCodeSuccess)
}

func handleVerificationSubmitCodeRequest(client *clients.Client, action *actions.Action) {
	code, ok := action.Payload["code"].(string)
	if !ok {
		client.WriteFailure(
			actions.VerificationSubmitCodeFailure,
			[]string{"code"},
		)
		return
	}

	isVerified, err := client.IsVerified()
	if err != nil {
		utils.LogBody(v2Tag, "error checking if client is verified. "+err.Error())
		client.WriteFailure(
			actions.VerificationSubmitCodeFailure,
			[]string{"server error"},
		)
		return
	}

	if isVerified {
		verificationCode, err := client.GetVerificationCode()
		if err != nil {
			utils.LogBody(v2Tag, "error getting client verification code. "+err.Error())
			client.WriteFailure(
				actions.VerificationSubmitCodeFailure,
				[]string{"server error"},
			)
			return
		}

		if code == verificationCode {
			client.WriteEmptyAction(actions.VerificationSubmitCodeSuccess)
		} else {
			client.WriteFailure(
				actions.VerificationSubmitCodeFailure,
				[]string{"wrong verification code"},
			)
		}

		return
	}

	clientRow, err := db.ReadClientByID(client.GetID())
	if err != nil {
		utils.LogBody(v2Tag, "error reading client by id. "+err.Error())
		client.WriteFailure(
			actions.VerificationSubmitCodeFailure,
			[]string{"server error"},
		)
		return
	}

	resp, err := api.GetTwilioVerificationCheck(clientRow.CountryCode, clientRow.PhoneNumber, code)
	if err != nil {
		utils.LogBody(v2Tag, "twilio verification check error. "+err.Error())
		client.WriteFailure(
			actions.VerificationSubmitCodeFailure,
			[]string{"server error"},
		)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		utils.LogBody(v2Tag, "error reading response body. "+err.Error())
		client.WriteFailure(
			actions.VerificationSubmitCodeFailure,
			[]string{"server error"},
		)
		return
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)
	if err != nil {
		utils.LogBody(v2Tag, "error parsing json. "+err.Error())
		client.WriteFailure(
			actions.VerificationSubmitCodeFailure,
			[]string{"server error"},
		)
		return
	}

	if !r["success"].(bool) {
		client.WriteFailure(
			actions.VerificationSubmitCodeFailure,
			[]string{"incorrect verification code"},
		)
		return
	}

	err = db.UpdateClientVerificationCode(client.GetID(), code)
	if err != nil {
		utils.LogBody(v2Tag, "error setting client as verified. "+err.Error())
		client.WriteFailure(
			actions.VerificationSubmitCodeFailure,
			[]string{"server error"},
		)
		return
	}

	client.WriteEmptyAction(actions.VerificationSubmitCodeSuccess)
}
