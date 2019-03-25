package v2

import (
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/clients"
	"github.com/hooligram/hooligram-server/db"
	"github.com/hooligram/hooligram-server/delivery"
	"github.com/hooligram/hooligram-server/utils"
)

func handleGroupAddMemberRequest(client *clients.Client, action *actions.Action) *actions.Action {
	if !client.IsSignedIn() {
		utils.LogBody(v2Tag, "client not signed in")
		failure := actions.GroupAddMemberFailure([]string{"not signed in"})
		client.WriteJSON(failure)
		return failure
	}

	groupID, ok := action.Payload["group_id"].(float64)
	if !ok {
		failure := actions.GroupAddMemberFailure([]string{"group_id not in payload"})
		client.WriteJSON(failure)
		return failure
	}

	newMemberID, ok := action.Payload["member_id"].(float64)
	if !ok {
		failure := actions.GroupAddMemberFailure([]string{"member_id not in payload"})
		client.WriteJSON(failure)
		return failure
	}

	if !db.ReadIsClientInMessageGroup(client.GetID(), int(groupID)) {
		failure := actions.GroupAddMemberFailure([]string{"not allowed"})
		client.WriteJSON(failure)
		return failure
	}

	err := db.CreateMessageGroupMembers(int(groupID), []int{int(newMemberID)})
	if err != nil {
		utils.LogBody(v2Tag, "error adding new message group member. "+err.Error())
		failure := actions.GroupAddMemberFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	memberIDs, err := db.ReadMessageGroupMemberIDs(int(groupID))
	if err != nil {
		utils.LogBody(v2Tag, "error reading message group member ids. "+err.Error())
		failure := actions.GroupAddMemberFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
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
		failure := actions.GroupAddMemberFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
	}

	messageGroupDelivery := delivery.MessageGroupDelivery{
		MessageGroup: messageGroup,
		RecipientIDs: recipientIDs,
	}
	delivery.GetMessageGroupDeliveryChan() <- &messageGroupDelivery

	success := actions.GroupAddMemberSuccess()
	client.WriteJSON(success)
	return success
}

func handleGroupCreateRequest(client *clients.Client, action *actions.Action) *actions.Action {
	groupName, ok := action.Payload["group_name"].(string)
	if !ok {
		failure := actions.GroupCreateFailure([]string{"group_name not in payload"})
		client.WriteJSON(failure)
		return failure
	}

	memberIDsPayload, ok := action.Payload["member_ids"].([]interface{})
	if !ok {
		failure := actions.GroupCreateFailure([]string{"member_ids not in payload"})
		client.WriteJSON(failure)
		return failure
	}

	memberIDs := make([]int, len(memberIDsPayload))

	for i, memberID := range memberIDsPayload {
		memberIDs[i] = int(memberID.(float64))
	}

	if len(memberIDs) < 2 {
		failure := actions.GroupCreateFailure([]string{"need at least two members"})
		client.WriteJSON(failure)
		return failure
	}

	if !utils.ContainsID(memberIDs, client.GetID()) {
		failure := actions.GroupCreateFailure([]string{"include group creator in member_ids"})
		client.WriteJSON(failure)
		return failure
	}

	messageGroup, err := db.CreateMessageGroup(groupName, memberIDs)
	if err != nil {
		utils.LogBody(v2Tag, "error creating message group. "+err.Error())
		failure := actions.GroupCreateFailure([]string{"server error"})
		client.WriteJSON(failure)
		return failure
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
	client.WriteJSON(success)
	return success
}

func handleGroupDeliverSuccess(client *clients.Client, action *actions.Action) *actions.Action {
	_, ok := action.Payload["message_group_id"].(float64)
	if !ok {
		failure := actions.GroupDeliverSuccessFailure([]string{"message_group_id not in payload"})
		client.WriteJSON(failure)
		return failure
	}

	success := actions.GroupDeliverSuccessSuccess()
	client.WriteJSON(success)
	return success
}

func handleGroupLeaveRequest(client *clients.Client, action *actions.Action) *actions.Action {
	if !client.IsSignedIn() {
		utils.LogBody(v2Tag, "client not signed in")
		failure := actions.GroupLeaveFailure([]string{"not signed in"})
		client.WriteJSON(failure)
		return failure
	}

	groupID, ok := action.Payload["group_id"].(float64)
	if !ok {
		failure := actions.GroupLeaveFailure(([]string{"group_id not in payload"}))
		client.WriteJSON(failure)
		return failure
	}

	if !db.ReadIsClientInMessageGroup(client.GetID(), int(groupID)) {
		failure := actions.GroupLeaveFailure(([]string{"not in group"}))
		client.WriteJSON(failure)
		return failure
	}

	err := db.DeleteMessageGroupMembers(int(groupID), []int{client.GetID()})
	if err != nil {
		utils.LogBody(v2Tag, "error removing client from message group. "+err.Error())
		failure := actions.GroupLeaveFailure(([]string{"server error"}))
		client.WriteJSON(failure)
		return failure
	}

	success := actions.GroupLeaveSuccess()
	client.WriteJSON(success)
	return success
}
