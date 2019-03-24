package v2

import (
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/clients"
	"github.com/hooligram/hooligram-server/db"
	"github.com/hooligram/hooligram-server/utils"
)

func handleAuthorizationSignInRequest(
	client *clients.Client,
	action *actions.Action,
) *actions.Action {
	countryCode, ok := action.Payload["country_code"].(string)
	if !ok {
		failure := actions.CreateAuthorizationSignInFailure(
			[]string{"country_code not in payload"},
		)
		client.WriteJSON(failure)
		return failure
	}

	phoneNumber, ok := action.Payload["phone_number"].(string)
	if !ok {
		failure := actions.CreateAuthorizationSignInFailure(
			[]string{"phone_number not in payload"},
		)
		client.WriteJSON(failure)
		return failure
	}

	verificationCode, ok := action.Payload["code"].(string)
	if !ok {
		failure := actions.CreateAuthorizationSignInFailure([]string{"code not in payload"})
		client.WriteJSON(failure)
		return failure
	}

	clientRow, err := db.ReadClientByUniqueKey(countryCode, phoneNumber)
	if err != nil {
		utils.LogBody(v2Tag, "error reading client by unique key. "+err.Error())
		failure := actions.CreateAuthorizationSignInFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	if clientRow == nil || clientRow.VerificationCode != verificationCode {
		failure := actions.CreateAuthorizationSignInFailure([]string{"not verified"})
		client.WriteJSON(failure)
		return failure
	}

	ok, err = client.SignIn(countryCode, phoneNumber, verificationCode)
	if err != nil {
		utils.LogBody(v2Tag, "error signing in client. "+err.Error())
		failure := actions.CreateAuthorizationSignInFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	if !ok {
		failure := actions.CreateAuthorizationSignInFailure([]string{"wrong credentials"})
		client.WriteJSON(failure)
		return failure
	}

	action.Type = actions.AuthorizationSignInSuccess
	client.WriteJSON(action)

	undeliveredMessages, err := db.FindUndeliveredMessages(client.GetID())
	if err != nil {
		utils.LogBody(v2Tag, "error finding messages to deliver. "+err.Error())
	}

	for _, undeliveredMessage := range undeliveredMessages {
		action := actions.CreateMessagingDeliverRequest(undeliveredMessage)
		client.WriteJSON(action)
	}

	return action
}