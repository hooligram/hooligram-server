package actions

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
