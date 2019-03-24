package actions

import "github.com/hooligram/hooligram-server/constants"

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
func GroupCreateSuccess(
	groupID int64,
	groupName string,
	memberIDs []int,
	dateCreated string,
) *Action {
	payload := make(map[string]interface{})
	memberIDs = append([]int(nil), memberIDs...)

	payload["id"] = groupID
	payload["date_created"] = dateCreated
	payload["member_ids"] = memberIDs
	payload["name"] = groupName

	return &Action{
		Payload: payload,
		Type:    constants.GroupCreateSuccess,
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
