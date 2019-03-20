package main

func createGroupAddMemberFailureAction(errors []string) *Action {
	payload := make(map[string]interface{})
	payload["errors"] = errors

	return &Action{
		Payload: payload,
		Type:    groupAddMemberFailure,
	}
}

func createGroupAddMemberSuccessAction() *Action {
	payload := make(map[string]interface{})

	return &Action{
		Payload: payload,
		Type:    groupAddMemberSuccess,
	}
}
