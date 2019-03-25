package actions

import (
	"github.com/hooligram/hooligram-server/constants"
	"github.com/hooligram/hooligram-server/db"
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
func GroupCreateSuccess(groupID int64) *Action {
	payload := make(map[string]interface{})
	payload["message_group_id"] = groupID

	return &Action{
		Payload: payload,
		Type:    constants.GroupCreateSuccess,
	}
}

///////////////////
// GROUP_DELIVER //
///////////////////

// GroupDeliverRequest .
func GroupDeliverRequest(messageGroup *db.MessageGroup) *Action {
	payload := make(map[string]interface{})
	payload["date_created"] = messageGroup.DateCreated
	payload["group_name"] = messageGroup.Name
	payload["member_ids"] = messageGroup.MemberIDs
	payload["message_group_id"] = messageGroup.ID

	return &Action{
		Payload: payload,
		Type:    constants.GroupDeliverRequest,
	}
}

// GroupDeliverSuccessFailure .
func GroupDeliverSuccessFailure(errors []string) *Action {
	return constructFailureAction(constants.GroupDeliverSuccessFailure, errors)
}

// GroupDeliverSuccessSuccess .
func GroupDeliverSuccessSuccess() *Action {
	return constructEmptyAction(constants.GroupDeliverSuccessSuccess)
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
