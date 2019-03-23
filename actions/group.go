package actions

//////////////////////
// GROUP_ADD_MEMBER //
//////////////////////

// CreateGroupAddMemberFailure .
func CreateGroupAddMemberFailure(errors []string) *Action {
	return createFailureAction(GroupAddMemberFailure, errors)
}

// CreateGroupAddMemberSuccess .
func CreateGroupAddMemberSuccess() *Action {
	return createEmptyAction(GroupAddMemberSuccess)
}

//////////////////
// GROUP_CREATE //
//////////////////

// CreateGroupCreateFailure .
func CreateGroupCreateFailure(errors []string) *Action {
	return createFailureAction(GroupCreateFailure, errors)
}

// CreateGroupCreateSuccess .
func CreateGroupCreateSuccess(
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
		Type:    GroupCreateSuccess,
	}
}

/////////////////
// GROUP_LEAVE //
/////////////////

// CreateGroupLeaveFailure .
func CreateGroupLeaveFailure(errors []string) *Action {
	return createFailureAction(GroupLeaveFailure, errors)
}

// CreateGroupLeaveSuccess .
func CreateGroupLeaveSuccess() *Action {
	return createEmptyAction(GroupLeaveSuccess)
}
