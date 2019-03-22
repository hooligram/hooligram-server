package main

func createEmptyAction(actionType string) *Action {
	payload := make(map[string]interface{})

	return &Action{
		Payload: payload,
		Type:    actionType,
	}
}

func createFailureAction(actionType string, errors []string) *Action {
	payload := make(map[string]interface{})
	payload["errors"] = errors

	return &Action{
		Payload: payload,
		Type:    actionType,
	}
}

func createGroupAddMemberFailureAction(errors []string) *Action {
	return createFailureAction(groupAddMemberFailure, errors)
}

func createGroupAddMemberSuccessAction() *Action {
	return createEmptyAction(groupAddMemberSuccess)
}

func createGroupLeaveFailureAction(errors []string) *Action {
	return createFailureAction(groupLeaveFailure, errors)
}

func createGroupLeaveSuccessAction() *Action {
	return createEmptyAction(groupLeaveSuccess)
}
