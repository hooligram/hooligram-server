package v2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/api"
	"github.com/hooligram/hooligram-server/constants"
	"github.com/hooligram/hooligram-server/db"
	"github.com/hooligram/hooligram-server/globals"
	"github.com/hooligram/hooligram-server/session"
	"github.com/hooligram/hooligram-server/structs"
	"github.com/hooligram/hooligram-server/utils"
)

const v2Tag = "v2"

func V2(w http.ResponseWriter, r *http.Request) {
	conn, err := globals.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.LogInfo(v2Tag, "websocket upgrade failed. "+err.Error())
		return
	}

	session.Add(conn)
	defer session.Remove(conn)
	defer conn.Close()

	for {
		client, ok := session.Get(conn)
		if !ok {
			break
		}

		var p []byte
		_, p, err = conn.ReadMessage()
		if err != nil {
			utils.LogInfo(
				v2Tag,
				fmt.Sprintf("connection error. client id %v. %v", client.ID, err.Error()),
			)
			return
		}

		action := structs.Action{}
		err = json.Unmarshal(p, &action)

		if err != nil {
			utils.LogInfo(
				v2Tag,
				fmt.Sprintf("error reading json. client id %v. %v", client.ID, err.Error()),
			)
			continue
		}

		if action.Type == "" {
			utils.LogInfo(
				v2Tag,
				fmt.Sprintf("action type missing. client id %v. %v", client.ID, err.Error()),
			)
			continue
		}

		if action.Payload == nil {
			utils.LogInfo(
				v2Tag,
				fmt.Sprintf("action payload missing. client id %v. %v", client.ID, err.Error()),
			)
			continue
		}

		utils.LogOpen(client, &action)
		var result *structs.Action

		switch action.Type {
		case constants.AuthorizationSignInRequest:
			result = handleAuthorizationSignInRequest(conn, &action)
		case constants.GroupAddMemberRequest:
			result = handleGroupAddMemberRequest(conn, &action)
		case constants.GroupCreateRequest:
			handleGroupCreateRequest(conn, &action)
		case constants.GroupLeaveRequest:
			result = handleGroupLeaveRequest(conn, &action)
		case constants.MessagingSendRequest:
			handleMessagingSendRequest(conn, &action)
		case constants.MessagingDeliverSuccess:
			handleMessagingDeliverSuccess(conn, &action)
		case constants.VerificationRequestCodeRequest:
			handleVerificationRequestCodeRequest(conn, &action)
		case constants.VerificationSubmitCodeRequest:
			handleVerificationSubmitCodeRequest(conn, &action)
		default:
		}

		if result == nil {
			continue
		}

		utils.LogClose(client, result)
	}
}

func handleAuthorizationSignInRequest(conn *websocket.Conn, action *structs.Action) *structs.Action {
	countryCode := action.Payload["country_code"].(string)
	phoneNumber := action.Payload["phone_number"].(string)
	verificationCode := action.Payload["code"].(string)

	client := session.SignIn(conn, countryCode, phoneNumber, verificationCode)
	if client == nil {
		client.WriteFailure(constants.AuthorizationSignInFailure, []string{"sign in failed"})
		return &structs.Action{
			Payload: map[string]interface{}{
				"errors": []string{"sign in failed"},
			},
			Type: constants.AuthorizationSignInFailure,
		}
	}

	action.Type = constants.AuthorizationSignInSuccess
	client.WriteJSON(action)

	undeliveredMessages, err := db.FindUndeliveredMessages(client.ID)
	if err != nil {
		utils.LogBody(v2Tag, "error finding messages to deliver. "+err.Error())
	}

	for _, undeliveredMessage := range undeliveredMessages {
		action := utils.ConstructDeliverMessageAction(undeliveredMessage)
		client.WriteJSON(action)
	}

	return action
}

func handleGroupAddMemberRequest(conn *websocket.Conn, action *structs.Action) *structs.Action {
	client, ok := session.Get(conn)
	if !ok {
		utils.LogBody(v2Tag, "error retrieving client")
		return nil
	}

	if !client.IsSignedIn {
		utils.LogBody(v2Tag, "client not signed in")
		failure := actions.CreateGroupAddMemberFailureAction([]string{"not signed in"})
		client.WriteJSON(failure)
		return failure
	}

	groupID, ok := action.Payload["group_id"].(float64)
	if !ok {
		failure := actions.CreateGroupAddMemberFailureAction([]string{"group_id is missing"})
		client.WriteJSON(failure)
		return failure
	}

	memberID, ok := action.Payload["member_id"].(float64)
	if !ok {
		failure := actions.CreateGroupAddMemberFailureAction([]string{"member_id is missing"})
		client.WriteJSON(failure)
		return failure
	}

	if !db.IsClientInMessageGroup(client.ID, int(groupID)) {
		failure := actions.CreateGroupAddMemberFailureAction([]string{"not allowed"})
		client.WriteJSON(failure)
		return failure
	}

	err := db.CreateMessageGroupMembers(int(groupID), []int{int(memberID)})
	if err != nil {
		utils.LogBody(v2Tag, "error adding new message group member. "+err.Error())
		failure := actions.CreateGroupAddMemberFailureAction([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	success := actions.CreateGroupAddMemberSuccessAction()
	client.WriteJSON(success)
	return success
}

func handleGroupCreateRequest(conn *websocket.Conn, action *structs.Action) {
	errors := []string{}

	client, ok := session.Get(conn)
	if !ok {
		return
	}

	groupName, groupNameOk := action.Payload["name"].(string)
	memberIDsPayload, memberIDsOk := action.Payload["member_ids"].([]interface{})
	memberIDs := make([]int, len(memberIDsPayload))

	for i, memberID := range memberIDsPayload {
		memberIDs[i] = int(memberID.(float64))
	}

	if !groupNameOk {
		errors = append(errors, "you need to include `name` in payload")
	}

	if !memberIDsOk {
		errors = append(errors, "you need to include `member_ids` in payload")
	}

	if len(memberIDs) < 1 {
		errors = append(
			errors,
			"you need to include at least one member in `member_ids` in payload",
		)
	}

	if !utils.ContainsID(memberIDs, client.ID) {
		errors = append(
			errors,
			"you need to include at the group creator in `member_ids` in payload",
		)
	}

	if len(errors) > 0 {
		errorText := ""

		for _, err := range errors {
			errorText += " " + err
		}

		utils.LogInfo(v2Tag, errorText)
		client.WriteFailure(constants.GroupCreateFailure, errors)
		return
	}

	messageGroup, err := db.CreateMessageGroup(groupName, memberIDs)
	if err != nil {
		client.WriteFailure(constants.GroupCreateFailure, errors)
	}

	successAction := utils.ConstructCreateGroupSuccessAction(
		messageGroup.ID,
		messageGroup.Name,
		messageGroup.MemberIDs,
		messageGroup.DateCreated,
	)

	client.WriteJSON(successAction)
}

func handleGroupLeaveRequest(conn *websocket.Conn, action *structs.Action) *structs.Action {
	client, ok := session.Get(conn)
	if !ok {
		utils.LogBody(v2Tag, "error retrieving client")
		return nil
	}

	if !client.IsSignedIn {
		utils.LogBody(v2Tag, "client not signed in")
		failure := actions.CreateGroupLeaveFailureAction([]string{"not signed in"})
		client.WriteJSON(failure)
		return failure
	}

	groupID, ok := action.Payload["group_id"].(float64)
	if !ok {
		failure := actions.CreateGroupLeaveFailureAction(([]string{"group_id is missing"}))
		client.WriteJSON(failure)
		return failure
	}

	if !db.IsClientInMessageGroup(client.ID, int(groupID)) {
		failure := actions.CreateGroupLeaveFailureAction(([]string{"not in group"}))
		client.WriteJSON(failure)
		return failure
	}

	err := db.DeleteMessageGroupMembers(int(groupID), []int{client.ID})
	if err != nil {
		utils.LogBody(v2Tag, "error removing client from message group. "+err.Error())
		failure := actions.CreateGroupLeaveFailureAction(([]string{"server error"}))
		client.WriteJSON(failure)
		return failure
	}

	success := actions.CreateGroupLeaveSuccessAction()
	client.WriteJSON(success)
	return success
}

func handleMessagingSendRequest(conn *websocket.Conn, action *structs.Action) {
	client, ok := session.Get(conn)
	if !ok {
		utils.LogBody(v2Tag, "error retrieving client")
	}

	content, ok := action.Payload["content"].(string)
	if !ok {
		client.WriteFailure(constants.MessagingSendFailure, []string{"content is missing"})
		return
	}

	messageGroupID, ok := action.Payload["message_group_id"].(float64)
	if !ok {
		client.WriteFailure(constants.MessagingSendFailure, []string{"message_group_id is missing"})
		return
	}

	senderID, ok := action.Payload["sender_id"].(float64)
	if !ok {
		client.WriteFailure(constants.MessagingSendFailure, []string{"sender_id is missing"})
		return
	}

	if client.ID != int(senderID) {
		client.WriteFailure(constants.MessagingSendFailure, []string{"sender_id mismatch"})
		return
	}

	if !db.IsClientInMessageGroup(int(senderID), int(messageGroupID)) {
		client.WriteFailure(constants.MessagingSendFailure, []string{"sender doesn't belong to message group"})
		return
	}

	message, err := db.CreateMessage(content, int(messageGroupID), int(senderID))
	if err != nil {
		client.WriteFailure(constants.MessagingSendFailure, []string{err.Error()})
		return
	}

	messageGroupMemberIDs, err := db.FindAllMessageGroupMemberIDs(message.MessageGroupID)
	if err != nil {
		client.WriteFailure(constants.MessagingSendFailure, []string{err.Error()})
		return
	}

	var recipientIDs []int

	for _, recipientID := range messageGroupMemberIDs {
		if recipientID == int(message.SenderID) {
			continue
		}

		recipientIDs = append(recipientIDs, recipientID)
	}

	for _, recipientID := range recipientIDs {
		db.CreateReceipt(message.ID, recipientID)
	}

	globals.MessageDeliveryChan <- &structs.MessageDelivery{
		Message:      message,
		RecipientIDs: recipientIDs,
	}

	payload := make(map[string]interface{})
	payload["message_id"] = message.ID
	client.WriteJSON(&structs.Action{
		Payload: payload,
		Type:    constants.MessagingSendSuccess,
	})
}

func handleMessagingDeliverSuccess(conn *websocket.Conn, action *structs.Action) {
	client, ok := session.Get(conn)
	if !ok {
		utils.LogBody(v2Tag, "error retrieving client")
	}

	messageID, ok := action.Payload["message_id"].(float64)
	if !ok {
		client.WriteFailure(constants.MessagingDeliverFailure, []string{"message_id is missing"})
		return
	}

	recipientID := client.ID
	err := db.UpdateReceiptDateDelivered(int(messageID), recipientID)
	if err != nil {
		client.WriteFailure(constants.MessagingDeliverFailure, []string{err.Error()})
		return
	}
}

func handleVerificationRequestCodeRequest(conn *websocket.Conn, action *structs.Action) {
	client, ok := session.Get(conn)
	if !ok {
		utils.LogBody(v2Tag, "error retrieving client")
	}

	errors := []string{}
	countryCode, ok := action.Payload["country_code"].(string)

	if !ok {
		errors = append(errors, "you need to include country_code in payload")
	}

	phoneNumber, ok := action.Payload["phone_number"].(string)

	if !ok {
		errors = append(errors, "you need to include phone_number in payload")
	}

	if len(errors) > 0 {
		client.WriteFailure(constants.VerificationRequestCodeFailure, errors)
		return
	}

	resp, err := api.PostTwilioVerificationStart(countryCode, phoneNumber)

	if err != nil {
		errors = append(errors, err.Error())
		client.WriteFailure(constants.VerificationRequestCodeFailure, errors)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err != nil {
		errors = append(errors, err.Error())
		client.WriteFailure(constants.VerificationRequestCodeFailure, errors)
		return
	}

	r := map[string]interface{}{}
	err = json.Unmarshal(body, &r)

	if err != nil {
		errors = append(errors, err.Error())
		client.WriteFailure(constants.VerificationRequestCodeFailure, errors)
		return
	}

	if !r["success"].(bool) {
		errors = append(errors, "i failed to make verification start api call")
		client.WriteFailure(constants.VerificationRequestCodeFailure, errors)
		return
	}

	client.WriteEmptyAction(constants.VerificationRequestCodeSuccess)
}

func handleVerificationSubmitCodeRequest(conn *websocket.Conn, action *structs.Action) {
	var errors []string
	client, ok := session.Get(conn)
	if !ok {
		utils.LogBody(v2Tag, "error retrieving client")
	}
	code, ok := action.Payload["code"].(string)

	if !ok {
		errors = append(errors, "you need to include code in payload")
		client.WriteFailure(constants.VerificationSubmitCodeFailure, errors)
		return
	}

	if !client.IsVerified() {
		resp, err := api.GetTwilioVerificationCheck(client.CountryCode, client.PhoneNumber, code)

		if err != nil {
			errors = append(errors, err.Error())
			client.WriteFailure(constants.VerificationSubmitCodeFailure, errors)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			errors = append(errors, err.Error())
			client.WriteFailure(constants.VerificationSubmitCodeFailure, errors)
			return
		}

		r := map[string]interface{}{}
		err = json.Unmarshal(body, &r)

		if err != nil {
			errors = append(errors, err.Error())
			client.WriteFailure(constants.VerificationSubmitCodeFailure, errors)
			return
		}

		if !r["success"].(bool) {
			// unverifyClient(client, conn)
			errors = append(errors, "twilio verification check failed")
			client.WriteFailure(constants.VerificationSubmitCodeFailure, errors)
			return
		}
	} else {
		if client.VerificationCode != code {
			// unverifyClient(client, conn)
			errors = append(errors, "verification code doesn't match")
			client.WriteFailure(constants.VerificationSubmitCodeFailure, errors)
			return
		}
	}

	// verifyClient(client, conn, code)
	client.WriteEmptyAction(constants.VerificationSubmitCodeSuccess)
}
