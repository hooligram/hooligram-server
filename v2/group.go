package v2

import (
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/clients"
	"github.com/hooligram/hooligram-server/db"
	"github.com/hooligram/hooligram-server/delivery"
	"github.com/hooligram/hooligram-server/utils"
)

///////////////////////////////////////
// HANDLER: GROUP_ADD_MEMBER_REQUEST //
///////////////////////////////////////

func handleGroupAddMemberRequest(client *clients.Client, action *actions.Action) *actions.Action {
	requestID := action.ID
	if requestID == "" {
		return groupAddMemberFailure(client, requestID, "id not in action")
	}

	if !client.IsSignedIn() {
		utils.LogBody(v2Tag, "client not signed in")
		return groupAddMemberFailure(client, requestID, "not signed in")
	}

	groupID, ok := action.Payload["group_id"].(float64)
	if !ok {
		return groupAddMemberFailure(client, requestID, "group_id not in payload")
	}

	newMemberID, ok := action.Payload["member_id"].(float64)
	if !ok {
		return groupAddMemberFailure(client, requestID, "member_id not in payload")
	}

	if !db.ReadIsClientInMessageGroup(client.GetID(), int(groupID)) {
		return groupAddMemberFailure(client, requestID, "not allowed")
	}

	err := db.CreateMessageGroupMembers(int(groupID), []int{int(newMemberID)})
	if err != nil {
		utils.LogBody(v2Tag, "error adding new message group member. "+err.Error())
		return groupAddMemberFailure(client, requestID, "server error")
	}

	memberIDs, err := db.ReadMessageGroupMemberIDs(int(groupID))
	if err != nil {
		utils.LogBody(v2Tag, "error reading message group member ids. "+err.Error())
		return groupAddMemberFailure(client, requestID, "server error")
	}

	recipientIDs := []int{}

	for _, memberID := range memberIDs {
		if memberID == client.GetID() {
			continue
		}

		recipientIDs = append(recipientIDs, memberID)
	}

	messageGroup, err := db.ReadMessageGroupByID(int(groupID))
	if err != nil {
		utils.LogBody(v2Tag, "error reading message group. "+err.Error())
		return groupAddMemberFailure(client, requestID, "server error")
	}

	messageGroupDelivery := delivery.MessageGroupDelivery{
		MessageGroup: messageGroup,
		RecipientIDs: recipientIDs,
	}
	delivery.GetMessageGroupDeliveryChan() <- &messageGroupDelivery

	success := actions.GroupAddMemberSuccess()
	success.ID = generateRandomActionID()
	client.WriteJSON(success)
	return success
}

///////////////////////////////////
// HANDLER: GROUP_CREATE_REQUEST //
///////////////////////////////////

func handleGroupCreateRequest(client *clients.Client, action *actions.Action) *actions.Action {
	requestID := action.ID
	if requestID == "" {
		return groupCreateFailure(client, requestID, "id not in action")
	}

	groupName, ok := action.Payload["group_name"].(string)
	if !ok {
		return groupCreateFailure(client, requestID, "group_name not in payload")
	}

	memberIDsPayload, ok := action.Payload["member_ids"].([]interface{})
	if !ok {
		return groupCreateFailure(client, requestID, "member_ids not in payload")
	}

	memberIDs := make([]int, len(memberIDsPayload))

	for i, memberID := range memberIDsPayload {
		memberIDs[i] = int(memberID.(float64))
	}

	if len(memberIDs) < 2 {
		return groupCreateFailure(client, requestID, "need at least two members")
	}

	if !utils.ContainsID(memberIDs, client.GetID()) {
		return groupCreateFailure(client, requestID, "include group creator in member_ids")
	}

	messageGroup, err := db.CreateMessageGroup(groupName, memberIDs)
	if err != nil {
		utils.LogBody(v2Tag, "error creating message group. "+err.Error())
		return groupCreateFailure(client, requestID, "server error")
	}

	recipientIDs := []int{}

	for _, memberID := range memberIDs {
		if memberID == client.GetID() {
			continue
		}

		recipientIDs = append(recipientIDs, memberID)
	}

	delivery.GetMessageGroupDeliveryChan() <- &delivery.MessageGroupDelivery{
		MessageGroup: messageGroup,
		RecipientIDs: recipientIDs,
	}

	success := actions.GroupCreateSuccess(messageGroup.ID)
	success.ID = generateRandomActionID()
	client.WriteJSON(success)
	return success
}

////////////////////////////////////
// HANDLER: GROUP_DELIVER_SUCCESS //
////////////////////////////////////

func handleGroupDeliverSuccess(client *clients.Client, action *actions.Action) *actions.Action {
	requestID := action.ID
	if requestID == "" {
		return groupDeliverSuccessFailure(client, requestID, "id not in action")
	}

	_, ok := action.Payload["message_group_id"].(float64)
	if !ok {
		return groupDeliverSuccessFailure(client, requestID, "message_group_id not in payload")
	}

	success := actions.GroupDeliverSuccessSuccess()
	success.ID = generateRandomActionID()
	client.WriteJSON(success)
	return success
}

//////////////////////////////////
// HANDLER: GROUP_LEAVE_REQUEST //
//////////////////////////////////

func handleGroupLeaveRequest(client *clients.Client, action *actions.Action) *actions.Action {
	requestID := action.ID
	if requestID == "" {
		return groupLeaveFailure(client, requestID, "id not in action")
	}

	if !client.IsSignedIn() {
		utils.LogBody(v2Tag, "client not signed in")
		return groupLeaveFailure(client, requestID, "not signed in")
	}

	groupID, ok := action.Payload["group_id"].(float64)
	if !ok {
		return groupLeaveFailure(client, requestID, "group_id not in payload")
	}

	if !db.ReadIsClientInMessageGroup(client.GetID(), int(groupID)) {
		return groupLeaveFailure(client, requestID, "not in group")
	}

	err := db.DeleteMessageGroupMembers(int(groupID), []int{client.GetID()})
	if err != nil {
		utils.LogBody(v2Tag, "error removing client from message group. "+err.Error())
		return groupLeaveFailure(client, requestID, "server error")
	}

	memberIDs, err := db.ReadMessageGroupMemberIDs(int(groupID))
	if err != nil {
		utils.LogBody(v2Tag, "error reading message group member ids. "+err.Error())
		return groupLeaveFailure(client, requestID, "server error")
	}

	recipientIDs := []int{}

	for _, memberID := range memberIDs {
		if memberID == client.GetID() {
			continue
		}

		recipientIDs = append(recipientIDs, memberID)
	}

	messageGroup, err := db.ReadMessageGroupByID(int(groupID))
	if err != nil {
		utils.LogBody(v2Tag, "error reading message group. "+err.Error())
		return groupLeaveFailure(client, requestID, "server error")
	}

	messageGroupDelivery := delivery.MessageGroupDelivery{
		MessageGroup: messageGroup,
		RecipientIDs: recipientIDs,
	}
	delivery.GetMessageGroupDeliveryChan() <- &messageGroupDelivery

	success := actions.GroupLeaveSuccess()
	success.ID = generateRandomActionID()
	client.WriteJSON(success)
	return success
}

////////////
// HELPER //
////////////

func groupAddMemberFailure(client *clients.Client, actionID, err string) *actions.Action {
	failure := actions.GroupAddMemberFailure([]string{err})
	failure.ID = actionID
	client.WriteJSON(failure)
	return failure
}

func groupCreateFailure(client *clients.Client, actionID, err string) *actions.Action {
	failure := actions.GroupCreateFailure([]string{err})
	failure.ID = actionID
	client.WriteJSON(failure)
	return failure
}

func groupDeliverSuccessFailure(client *clients.Client, actionID, err string) *actions.Action {
	failure := actions.GroupDeliverSuccessFailure([]string{err})
	failure.ID = actionID
	client.WriteJSON(failure)
	return failure
}

func groupLeaveFailure(client *clients.Client, actionID, err string) *actions.Action {
	failure := actions.GroupLeaveFailure([]string{err})
	failure.ID = actionID
	client.WriteJSON(failure)
	return failure
}
