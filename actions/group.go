package actions

import (
	"github.com/hooligram/hooligram-server/constants"
	"github.com/hooligram/hooligram-server/db"
	"github.com/hooligram/hooligram-server/utils"
)

//////////////////////
// GROUP_ADD_MEMBER //
//////////////////////

// GroupAddMemberFailure .
func GroupAddMemberFailure(errors []string) *Action {
	return constructFailureAction(constants.GroupAddMemberFailure, errors)
}

// GroupAddMemberSuccess .
func GroupAddMemberSuccess() *Action {
	return constructEmptyAction(constants.GroupAddMemberSuccess)
}

//////////////////
// GROUP_CREATE //
//////////////////

// GroupCreateFailure .
func GroupCreateFailure(errors []string) *Action {
	return constructFailureAction(constants.GroupCreateFailure, errors)
}

// GroupCreateSuccess .
func GroupCreateSuccess(groupID int, actionID string) *Action {
	payload := make(map[string]interface{})
	payload["action_id"] = actionID
	payload["group_id"] = groupID

	return &Action{
		Payload: payload,
		Type:    constants.GroupCreateSuccess,
	}
}

///////////////////
// GROUP_DELIVER //
///////////////////

// GroupDeliverRequest .
func GroupDeliverRequest(messageGroupID int) *Action {
	messageGroup, err := db.ReadMessageGroupByID(messageGroupID)
	if err != nil {
		utils.LogInfo(actionsTag, "error reading message group by id. "+err.Error())
		return &Action{}
	}

	memberSIDs, err := messageGroup.MemberSIDs()
	if err != nil {
		utils.LogInfo(actionsTag, "error getting message group member sids. "+err.Error())
		return &Action{}
	}

	payload := make(map[string]interface{})
	payload["date_created"] = messageGroup.DateCreated
	payload["group_id"] = messageGroup.ID
	payload["group_name"] = messageGroup.Name
	payload["member_sids"] = memberSIDs

	return &Action{
		Payload: payload,
		Type:    constants.GroupDeliverRequest,
	}
}

/////////////////
// GROUP_LEAVE //
/////////////////

// GroupLeaveFailure .
func GroupLeaveFailure(errors []string) *Action {
	return constructFailureAction(constants.GroupLeaveFailure, errors)
}

// GroupLeaveSuccess .
func GroupLeaveSuccess() *Action {
	return constructEmptyAction(constants.GroupLeaveSuccess)
}
