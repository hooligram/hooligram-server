package v2

import (
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/clients"
	"github.com/hooligram/hooligram-server/db"
	"github.com/hooligram/hooligram-server/utils"
)

////////////////////////////////////////////
// HANDLER: AUTHORIZATION_SIGN_IN_REQUEST //
////////////////////////////////////////////

func handleAuthorizationSignInRequest(
	client *clients.Client,
	action *actions.Action,
) *actions.Action {
	actionID := action.ID
	if actionID == "" {
		return authorizationSignInFailure(client, actionID, "id not in action")
	}

	countryCode, ok := action.Payload["country_code"].(string)
	if !ok {
		return authorizationSignInFailure(client, actionID, "country_code not in payload")
	}

	phoneNumber, ok := action.Payload["phone_number"].(string)
	if !ok {
		return authorizationSignInFailure(client, actionID, "phone_number not in payload")
	}

	verificationCode, ok := action.Payload["verification_code"].(string)
	if !ok {
		return authorizationSignInFailure(client, actionID, "verification_code not in payload")
	}

	clientRow, ok, err := db.ReadClientByUniqueKey(countryCode, phoneNumber)
	if err != nil {
		utils.LogBody(v2Tag, "error reading client by unique key. "+err.Error())
		return authorizationSignInFailure(client, actionID, "server error")
	}

	if !ok || clientRow.VerificationCode != verificationCode {
		return authorizationSignInFailure(client, actionID, "not verified")
	}

	ok, err = client.SignIn(countryCode, phoneNumber, verificationCode)
	if err != nil {
		utils.LogBody(v2Tag, "error signing in client. "+err.Error())
		return authorizationSignInFailure(client, actionID, "server error")
	}

	if !ok {
		return authorizationSignInFailure(client, actionID, "wrong credentials")
	}

	success := actions.AuthorizationSignInSuccess(actionID)
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
		request := actions.GroupDeliverRequest(actionID, messageGroupID)
		client.WriteJSON(request)
	}

	for _, undeliveredMessage := range undeliveredMessages {
		request := actions.MessagingDeliverRequest(undeliveredMessage)
		client.WriteJSON(request)
	}

	return success
}

////////////
// HELPER //
////////////

func authorizationSignInFailure(client *clients.Client, actionID, err string) *actions.Action {
	failure := actions.AuthorizationSignInFailure(actionID, []string{err})
	client.WriteJSON(failure)
	return failure
}
