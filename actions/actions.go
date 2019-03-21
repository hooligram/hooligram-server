package actions

import (
	"github.com/hooligram/hooligram-server/constants"
	"github.com/hooligram/hooligram-server/structs"
)

// CreateGroupAddMemberFailureAction .
func CreateGroupAddMemberFailureAction(errors []string) *structs.Action {
	return createFailureAction(constants.GroupAddMemberFailure, errors)
}

// CreateGroupAddMemberSuccessAction .
func CreateGroupAddMemberSuccessAction() *structs.Action {
	return createEmptyAction(constants.GroupAddMemberSuccess)
}

// CreateGroupLeaveFailureAction .
func CreateGroupLeaveFailureAction(errors []string) *structs.Action {
	return createFailureAction(constants.GroupLeaveFailure, errors)
}

// CreateGroupLeaveSuccessAction .
func CreateGroupLeaveSuccessAction() *structs.Action {
	return createEmptyAction(constants.GroupLeaveSuccess)
}

func createEmptyAction(actionType string) *structs.Action {
	payload := make(map[string]interface{})

	return &structs.Action{
		Payload: payload,
		Type:    actionType,
	}
}

func createFailureAction(actionType string, errors []string) *structs.Action {
	payload := make(map[string]interface{})
	payload["errors"] = errors

	return &structs.Action{
		Payload: payload,
		Type:    actionType,
	}
}
