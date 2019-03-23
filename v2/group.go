package v2

import (
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/clients"
	"github.com/hooligram/hooligram-server/db"
	"github.com/hooligram/hooligram-server/utils"
)

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
		client.WriteFailure(actions.GroupCreateFailure, errors)
		return
	}

	messageGroup, err := db.CreateMessageGroup(groupName, memberIDs)
	if err != nil {
		client.WriteFailure(actions.GroupCreateFailure, errors)
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
