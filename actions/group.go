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
func GroupAddMemberFailure(actionID string, errors []string) *Action {
	return constructFailureAction(actionID, constants.GroupAddMemberFailure, errors)
}

// GroupAddMemberSuccess .
func GroupAddMemberSuccess(actionID string) *Action {
	return constructEmptyAction(actionID, constants.GroupAddMemberSuccess)
}

//////////////////
// GROUP_CREATE //
//////////////////

// GroupCreateFailure .
func GroupCreateFailure(actionID string, errors []string) *Action {
	return constructFailureAction(actionID, constants.GroupCreateFailure, errors)
}

// GroupCreateSuccess .
func GroupCreateSuccess(actionID string, groupID int) *Action {
	payload := make(map[string]interface{})
	payload["group_id"] = groupID

	return &Action{
		ID:      actionID,
		Payload: payload,
		Type:    constants.GroupCreateSuccess,
	}
}

///////////////////
// GROUP_DELIVER //
///////////////////

// GroupDeliverRequest .
func GroupDeliverRequest(actionID string, messageGroupID int) *Action {
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
	payload["group_type"] = messageGroup.Type
	payload["member_sids"] = memberSIDs

	return &Action{
		ID:      actionID,
		Payload: payload,
		Type:    constants.GroupDeliverRequest,
	}
}

/////////////////
// GROUP_LEAVE //
/////////////////

// GroupLeaveFailure .
func GroupLeaveFailure(actionID string, errors []string) *Action {
	return constructFailureAction(actionID, constants.GroupLeaveFailure, errors)
}

// GroupLeaveSuccess .
func GroupLeaveSuccess(actionID string) *Action {
	return constructEmptyAction(actionID, constants.GroupLeaveSuccess)
}
