package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/clients"
	"github.com/hooligram/hooligram-server/constants"
	"github.com/hooligram/hooligram-server/db"
	"github.com/hooligram/hooligram-server/globals"
	"github.com/hooligram/hooligram-server/utils"
)

const v2Tag = "v2"

// V2 .
func V2(w http.ResponseWriter, r *http.Request) {
	conn, err := globals.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		utils.LogInfo(v2Tag, "error upgrading to websocket. "+err.Error())
		return
	}

	clients.Add(conn)
	defer clients.Remove(conn)
	defer conn.Close()

	for {
		client, ok := clients.Get(conn)
		if !ok {
			break
		}

		var p []byte
		_, p, err = conn.ReadMessage()
		if err != nil {
			utils.LogInfo(
				v2Tag,
				fmt.Sprintf("connection error. client id %v. %v", client.GetID(), err.Error()),
			)
			return
		}

		action := actions.Action{}
		err = json.Unmarshal(p, &action)

		if err != nil {
			utils.LogInfo(
				v2Tag,
				fmt.Sprintf("error reading json. client id %v. %v", client.GetID(), err.Error()),
			)
			continue
		}

		if action.Type == "" {
			utils.LogInfo(
				v2Tag,
				fmt.Sprintf("action type missing. client id %v. %v", client.GetID(), err.Error()),
			)
			continue
		}

		if action.Payload == nil {
			utils.LogInfo(
				v2Tag,
				fmt.Sprintf("action payload missing. client id %v. %v", client.GetID(), err.Error()),
			)
			continue
		}

		utils.LogOpen(client.SessionID, strconv.Itoa(client.GetID()), action.Type, action.Payload)
		var result *actions.Action

		switch action.Type {
		case constants.AuthorizationSignInRequest:
			result = handleAuthorizationSignInRequest(client, &action)
		case constants.GroupAddMemberRequest:
			result = handleGroupAddMemberRequest(client, &action)
		case constants.GroupCreateRequest:
			handleGroupCreateRequest(client, &action)
		case constants.GroupLeaveRequest:
			result = handleGroupLeaveRequest(client, &action)
		case constants.MessagingSendRequest:
			handleMessagingSendRequest(client, &action)
		case constants.MessagingDeliverSuccess:
			handleMessagingDeliverSuccess(client, &action)
		case constants.VerificationRequestCodeRequest:
			handleVerificationRequestCodeRequest(client, &action)
		case constants.VerificationSubmitCodeRequest:
			handleVerificationSubmitCodeRequest(client, &action)
		default:
		}

		if result == nil {
			continue
		}

		utils.LogClose(client.SessionID, strconv.Itoa(client.GetID()), result.Type, result.Payload)
	}
}

func handleGroupAddMemberRequest(client *clients.Client, action *actions.Action) *actions.Action {
	if !client.IsSignedIn() {
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

	if !db.IsClientInMessageGroup(client.GetID(), int(groupID)) {
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

func handleGroupCreateRequest(client *clients.Client, action *actions.Action) {
	errors := []string{}

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

	if !utils.ContainsID(memberIDs, client.GetID()) {
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

	successAction := actions.CreateGroupCreateSuccessAction(
		messageGroup.ID,
		messageGroup.Name,
		messageGroup.MemberIDs,
		messageGroup.DateCreated,
	)

	client.WriteJSON(successAction)
}

func handleGroupLeaveRequest(client *clients.Client, action *actions.Action) *actions.Action {
	if !client.IsSignedIn() {
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

	if !db.IsClientInMessageGroup(client.GetID(), int(groupID)) {
		failure := actions.CreateGroupLeaveFailureAction(([]string{"not in group"}))
		client.WriteJSON(failure)
		return failure
	}

	err := db.DeleteMessageGroupMembers(int(groupID), []int{client.GetID()})
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

func handleMessagingSendRequest(client *clients.Client, action *actions.Action) {
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

	if client.GetID() != int(senderID) {
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

	globals.MessageDeliveryChan <- &globals.MessageDelivery{
		Message:      message,
		RecipientIDs: recipientIDs,
	}

	payload := make(map[string]interface{})
	payload["message_id"] = message.ID
	client.WriteJSON(&actions.Action{
		Payload: payload,
		Type:    constants.MessagingSendSuccess,
	})
}

func handleMessagingDeliverSuccess(client *clients.Client, action *actions.Action) {
	messageID, ok := action.Payload["message_id"].(float64)
	if !ok {
		client.WriteFailure(constants.MessagingDeliverFailure, []string{"message_id is missing"})
		return
	}

	recipientID := client.GetID()
	err := db.UpdateReceiptDateDelivered(int(messageID), recipientID)
	if err != nil {
		client.WriteFailure(constants.MessagingDeliverFailure, []string{err.Error()})
		return
	}
}
