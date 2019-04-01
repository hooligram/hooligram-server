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
	if !client.IsSignedIn() {
		return groupAddMemberFailure(client, "not signed in")
	}

	groupID, ok := action.Payload["group_id"].(float64)
	if !ok {
		return groupAddMemberFailure(client, "group_id not in payload")
	}

	newMemberSID, ok := action.Payload["member_sid"].(string)
	if !ok {
		return groupAddMemberFailure(client, "member_sid not in payload")
	}

	if !db.ReadIsClientInMessageGroup(client.GetID(), int(groupID)) {
		return groupAddMemberFailure(client, "not allowed")
	}

	newMemberRow, ok, err := db.ReadClientByUniqueKey(utils.ParseSID(newMemberSID))
	if err != nil {
		utils.LogBody(v2Tag, "error reading new member by unique key. "+err.Error())
		return groupAddMemberFailure(client, "server error")
	}

	if !ok {
		return groupAddMemberFailure(client, "new member not found")
	}

	err = db.CreateMessageGroupMembers(int(groupID), []int{int(newMemberRow.ID)})
	if err != nil {
		utils.LogBody(v2Tag, "error adding new message group member. "+err.Error())
		return groupAddMemberFailure(client, "server error")
	}

	memberIDs, err := db.ReadMessageGroupMemberIDs(int(groupID))
	if err != nil {
		utils.LogBody(v2Tag, "error reading message group member ids. "+err.Error())
		return groupAddMemberFailure(client, "server error")
	}

	messageGroup, err := db.ReadMessageGroupByID(int(groupID))
	if err != nil {
		utils.LogBody(v2Tag, "error reading message group. "+err.Error())
		return groupAddMemberFailure(client, "server error")
	}

	messageGroupDelivery := delivery.MessageGroupDelivery{
		MessageGroup: messageGroup,
		RecipientIDs: memberIDs,
	}
	delivery.GetMessageGroupDeliveryChan() <- &messageGroupDelivery

	success := actions.GroupAddMemberSuccess()
	client.WriteJSON(success)
	return success
}

///////////////////////////////////
// HANDLER: GROUP_CREATE_REQUEST //
///////////////////////////////////

func handleGroupCreateRequest(client *clients.Client, action *actions.Action) *actions.Action {
	if !client.IsSignedIn() {
		return groupCreateFailure(client, "not signed in")
	}

	groupName, ok := action.Payload["group_name"].(string)
	if !ok {
		return groupCreateFailure(client, "group_name not in payload")
	}

	memberSIDsPayload, ok := action.Payload["member_sids"].([]interface{})
	if !ok {
		return groupCreateFailure(client, "member_sids not in payload")
	}

	memberSIDs := make([]string, len(memberSIDsPayload))
	for i, memberSID := range memberSIDsPayload {
		memberSIDs[i] = memberSID.(string)
	}

	if len(memberSIDs) < 2 {
		return groupCreateFailure(client, "need at least two members")
	}

	if !utils.ContainsString(memberSIDs, client.GetSID()) {
		return groupCreateFailure(client, "include group creator in member_sids")
	}

	memberIDs := []int{}
	for _, memberSID := range memberSIDs {
		countryCode, phoneNumber := utils.ParseSID(memberSID)
		clientRow, ok, err := db.ReadClientByUniqueKey(countryCode, phoneNumber)
		if err != nil {
			utils.LogBody(v2Tag, "error reading client by unique key. "+err.Error())
			return groupCreateFailure(client, "server error")
		}

		if !ok {
			return groupCreateFailure(client, "member not found")
		}

		memberIDs = append(memberIDs, clientRow.ID)
	}

	messageGroup, err := db.CreateMessageGroup(groupName, memberIDs)
	if err != nil {
		utils.LogBody(v2Tag, "error creating message group. "+err.Error())
		return groupCreateFailure(client, "server error")
	}

	delivery.GetMessageGroupDeliveryChan() <- &delivery.MessageGroupDelivery{
		MessageGroup: messageGroup,
		RecipientIDs: memberIDs,
	}

	success := actions.GroupCreateSuccess(messageGroup.ID)
	client.WriteJSON(success)
	return success
}

//////////////////////////////////
// HANDLER: GROUP_LEAVE_REQUEST //
//////////////////////////////////

func handleGroupLeaveRequest(client *clients.Client, action *actions.Action) *actions.Action {
	if !client.IsSignedIn() {
		return groupLeaveFailure(client, "not signed in")
	}

	groupID, ok := action.Payload["group_id"].(float64)
	if !ok {
		return groupLeaveFailure(client, "group_id not in payload")
	}

	if !db.ReadIsClientInMessageGroup(client.GetID(), int(groupID)) {
		return groupLeaveFailure(client, "not in group")
	}

	err := db.DeleteMessageGroupMembers(int(groupID), []int{client.GetID()})
	if err != nil {
		utils.LogBody(v2Tag, "error removing client from message group. "+err.Error())
		return groupLeaveFailure(client, "server error")
	}

	memberIDs, err := db.ReadMessageGroupMemberIDs(int(groupID))
	if err != nil {
		utils.LogBody(v2Tag, "error reading message group member ids. "+err.Error())
		return groupLeaveFailure(client, "server error")
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
		return groupLeaveFailure(client, "server error")
	}

	messageGroupDelivery := delivery.MessageGroupDelivery{
		MessageGroup: messageGroup,
		RecipientIDs: recipientIDs,
	}
	delivery.GetMessageGroupDeliveryChan() <- &messageGroupDelivery

	success := actions.GroupLeaveSuccess()
	client.WriteJSON(success)
	return success
}

////////////
// HELPER //
////////////

func groupAddMemberFailure(client *clients.Client, err string) *actions.Action {
	failure := actions.GroupAddMemberFailure([]string{err})
	client.WriteJSON(failure)
	return failure
}

func groupCreateFailure(client *clients.Client, err string) *actions.Action {
	failure := actions.GroupCreateFailure([]string{err})
	client.WriteJSON(failure)
	return failure
}

func groupLeaveFailure(client *clients.Client, err string) *actions.Action {
	failure := actions.GroupLeaveFailure([]string{err})
	client.WriteJSON(failure)
	return failure
}
