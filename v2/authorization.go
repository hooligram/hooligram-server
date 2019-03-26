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
	requestID := action.ID
	if requestID == "" {
		failure := actions.AuthorizationSignInFailure([]string{"id not in action"})
		client.WriteJSON(failure)
		return failure
	}

	countryCode, ok := action.Payload["country_code"].(string)
	if !ok {
		failure := actions.AuthorizationSignInFailure(
			[]string{"country_code not in payload"},
		)
		failure.ID = requestID
		client.WriteJSON(failure)
		return failure
	}

	phoneNumber, ok := action.Payload["phone_number"].(string)
	if !ok {
		failure := actions.AuthorizationSignInFailure(
			[]string{"phone_number not in payload"},
		)
		failure.ID = requestID
		client.WriteJSON(failure)
		return failure
	}

	verificationCode, ok := action.Payload["verification_code"].(string)
	if !ok {
		failure := actions.AuthorizationSignInFailure([]string{"verification_code not in payload"})
		failure.ID = requestID
		client.WriteJSON(failure)
		return failure
	}

	clientRow, err := db.ReadClientByUniqueKey(countryCode, phoneNumber)
	if err != nil {
		utils.LogBody(v2Tag, "error reading client by unique key. "+err.Error())
		failure := actions.AuthorizationSignInFailure([]string{"server error"})
		failure.ID = requestID
		client.WriteJSON(failure)
		return failure
	}

	if clientRow == nil || clientRow.VerificationCode != verificationCode {
		failure := actions.AuthorizationSignInFailure([]string{"not verified"})
		failure.ID = requestID
		client.WriteJSON(failure)
		return failure
	}

	ok, err = client.SignIn(countryCode, phoneNumber, verificationCode)
	if err != nil {
		utils.LogBody(v2Tag, "error signing in client. "+err.Error())
		failure := actions.AuthorizationSignInFailure([]string{"server error"})
		failure.ID = requestID
		client.WriteJSON(failure)
		return failure
	}

	if !ok {
		failure := actions.AuthorizationSignInFailure([]string{"wrong credentials"})
		failure.ID = requestID
		client.WriteJSON(failure)
		return failure
	}

	success := actions.AuthorizationSignInSuccess()
	success.ID = requestID
	client.WriteJSON(success)

	undeliveredMessages, err := db.ReadUndeliveredMessages(client.GetID())
	if err != nil {
		utils.LogBody(v2Tag, "error finding messages to deliver. "+err.Error())
	}

	messageGroupIDs, err := db.ReadClientMessageGroupIDs(client.GetID())
	if err != nil {
		utils.LogBody(v2Tag, "error reading client message group ids. "+err.Error())
	}

	for _, messageGroupID := range messageGroupIDs {
		request := actions.GroupDeliverRequest(messageGroupID)
		request.ID = generateRandomActionID()
		client.WriteJSON(request)
	}

	for _, undeliveredMessage := range undeliveredMessages {
		request := actions.MessagingDeliverRequest(undeliveredMessage)
		request.ID = generateRandomActionID()
		client.WriteJSON(request)
	}

	return success
}
