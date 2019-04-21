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
	actionID := action.ID
	if actionID == "" {
		return groupAddMemberFailure(client, actionID, "id not in action")
	}

	if !client.IsSignedIn() {
		return groupAddMemberFailure(client, actionID, "not signed in")
	}

	groupID, ok := action.Payload["group_id"].(float64)
	if !ok {
		return groupAddMemberFailure(client, actionID, "group_id not in payload")
	}

	newMemberSID, ok := action.Payload["member_sid"].(string)
	if !ok {
		return groupAddMemberFailure(client, actionID, "member_sid not in payload")
	}

	if !db.ReadIsClientInMessageGroup(client.GetID(), int(groupID)) {
		return groupAddMemberFailure(client, actionID, "not allowed")
	}

	newMemberRow, ok, err := db.ReadClientByUniqueKey(utils.ParseSID(newMemberSID))
	if err != nil {
		utils.LogBody(v2Tag, "error reading new member by unique key. "+err.Error())
		return groupAddMemberFailure(client, actionID, "server error")
	}

	if !ok {
		return groupAddMemberFailure(client, actionID, "new member not found")
	}

	err = db.CreateMessageGroupMembers(int(groupID), []int{int(newMemberRow.ID)})
	if err != nil {
		utils.LogBody(v2Tag, "error adding new message group member. "+err.Error())
		return groupAddMemberFailure(client, actionID, "server error")
	}

	memberIDs, err := db.ReadMessageGroupMemberIDs(int(groupID))
	if err != nil {
		utils.LogBody(v2Tag, "error reading message group member ids. "+err.Error())
		return groupAddMemberFailure(client, actionID, "server error")
	}

	messageGroup, err := db.ReadMessageGroupByID(int(groupID))
	if err != nil {
		utils.LogBody(v2Tag, "error reading message group. "+err.Error())
		return groupAddMemberFailure(client, actionID, "server error")
	}

	messageGroupDelivery := delivery.MessageGroupDelivery{
		MessageGroup: messageGroup,
		RecipientIDs: memberIDs,
	}
	delivery.GetMessageGroupDeliveryChan() <- &messageGroupDelivery

	success := actions.GroupAddMemberSuccess(actionID)
	client.WriteJSON(success)
	return success
}

///////////////////////////////////
// HANDLER: GROUP_CREATE_REQUEST //
///////////////////////////////////

func handleGroupCreateRequest(client *clients.Client, action *actions.Action) *actions.Action {
	actionID := action.ID
	if actionID == "" {
		return groupCreateFailure(client, actionID, "id not in action")
	}

	if !client.IsSignedIn() {
		return groupCreateFailure(client, actionID, "not signed in")
	}

	groupName, ok := action.Payload["group_name"].(string)
	if !ok {
		return groupCreateFailure(client, actionID, "group_name not in payload")
	}

	memberSIDsPayload, ok := action.Payload["member_sids"].([]interface{})
	if !ok {
		return groupCreateFailure(client, actionID, "member_sids not in payload")
	}

	memberSIDs := make([]string, len(memberSIDsPayload))
	for i, memberSID := range memberSIDsPayload {
		memberSIDs[i] = memberSID.(string)
	}

	if len(memberSIDs) < 2 {
		return groupCreateFailure(client, actionID, "need at least two members")
	}

	if !utils.ContainsString(memberSIDs, client.GetSID()) {
		return groupCreateFailure(client, actionID, "include group creator in member_sids")
	}

	memberIDs := []int{}
	for _, memberSID := range memberSIDs {
		countryCode, phoneNumber := utils.ParseSID(memberSID)
		clientRow, ok, err := db.ReadClientByUniqueKey(countryCode, phoneNumber)
		if err != nil {
			utils.LogBody(v2Tag, "error reading client by unique key. "+err.Error())
			return groupCreateFailure(client, actionID, "server error")
		}

		if !ok {
			newClient, err := db.CreateClient(countryCode, phoneNumber)
			if err != nil {
				utils.LogBody(v2Tag, "error creating client. "+err.Error())
				return groupCreateFailure(client, actionID, "server error")
			}
			memberIDs = append(memberIDs, newClient.ID)
		} else {
			memberIDs = append(memberIDs, clientRow.ID)
		}

	}

	var messageGroup *db.MessageGroup

	if len(memberIDs) == 2 {
		groupID, err := db.ReadDirectMessageGroupID(memberIDs[0], memberIDs[1])
		if err != nil {
			utils.LogBody(v2Tag, "error reading direct message group id. "+err.Error())
			return groupCreateFailure(client, actionID, "server error")
		}

		if groupID == 0 {
			messageGroup, err = db.CreateMessageGroup(groupName, memberIDs)
			if err != nil {
				utils.LogBody(v2Tag, "error creating message group. "+err.Error())
				return groupCreateFailure(client, actionID, "server error")
			}

			err = db.CreateDirectMessage(messageGroup.ID, memberIDs[0], memberIDs[1])
			if err != nil {
				utils.LogBody(v2Tag, "error creating direct message. "+err.Error())
				return groupCreateFailure(client, actionID, "server error")
			}
		} else {
			messageGroup, err = db.ReadMessageGroupByID(groupID)
			if err != nil {
				utils.LogBody(v2Tag, "error reading message group by id. "+err.Error())
				return groupCreateFailure(client, actionID, "server error")
			}
		}
	} else {
		var err error
		messageGroup, err = db.CreateMessageGroup(groupName, memberIDs)
		if err != nil {
			utils.LogBody(v2Tag, "error creating message group. "+err.Error())
			return groupCreateFailure(client, actionID, "server error")
		}
	}

	delivery.GetMessageGroupDeliveryChan() <- &delivery.MessageGroupDelivery{
		MessageGroup: messageGroup,
		RecipientIDs: memberIDs,
	}

	success := actions.GroupCreateSuccess(actionID, messageGroup.ID)
	client.WriteJSON(success)
	return success
}

//////////////////////////////////
// HANDLER: GROUP_LEAVE_REQUEST //
//////////////////////////////////

func handleGroupLeaveRequest(client *clients.Client, action *actions.Action) *actions.Action {
	actionID := action.ID
	if actionID == "" {
		return groupLeaveFailure(client, actionID, "id not in action")
	}

	if !client.IsSignedIn() {
		return groupLeaveFailure(client, actionID, "not signed in")
	}

	groupID, ok := action.Payload["group_id"].(float64)
	if !ok {
		return groupLeaveFailure(client, actionID, "group_id not in payload")
	}

	if !db.ReadIsClientInMessageGroup(client.GetID(), int(groupID)) {
		return groupLeaveFailure(client, actionID, "not in group")
	}

	err := db.DeleteMessageGroupMembers(int(groupID), []int{client.GetID()})
	if err != nil {
		utils.LogBody(v2Tag, "error removing client from message group. "+err.Error())
		return groupLeaveFailure(client, actionID, "server error")
	}

	memberIDs, err := db.ReadMessageGroupMemberIDs(int(groupID))
	if err != nil {
		utils.LogBody(v2Tag, "error reading message group member ids. "+err.Error())
		return groupLeaveFailure(client, actionID, "server error")
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
		return groupLeaveFailure(client, actionID, "server error")
	}

	messageGroupDelivery := delivery.MessageGroupDelivery{
		MessageGroup: messageGroup,
		RecipientIDs: recipientIDs,
	}
	delivery.GetMessageGroupDeliveryChan() <- &messageGroupDelivery

	success := actions.GroupLeaveSuccess(actionID)
	client.WriteJSON(success)
	return success
}

////////////
// HELPER //
////////////

func groupAddMemberFailure(client *clients.Client, actionID, err string) *actions.Action {
	failure := actions.GroupAddMemberFailure(actionID, []string{err})
	client.WriteJSON(failure)
	return failure
}

func groupCreateFailure(client *clients.Client, actionID, err string) *actions.Action {
	failure := actions.GroupCreateFailure(actionID, []string{err})
	client.WriteJSON(failure)
	return failure
}

func groupLeaveFailure(client *clients.Client, actionID, err string) *actions.Action {
	failure := actions.GroupLeaveFailure(actionID, []string{err})
	client.WriteJSON(failure)
	return failure
}
