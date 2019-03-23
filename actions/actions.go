package actions

import (
	"github.com/hooligram/hooligram-server/constants"
	"github.com/hooligram/hooligram-server/db"
)

// Action .
type Action struct {
	Payload map[string]interface{} `json:"payload"`
	Type    string                 `json:"type"`
}

// CreateAuthorizationSignInFailureAction .
func CreateAuthorizationSignInFailureAction(errors []string) *Action {
	return createFailureAction(constants.AuthorizationSignInFailure, errors)
}

// CreateGroupAddMemberFailureAction .
func CreateGroupAddMemberFailureAction(errors []string) *Action {
	return createFailureAction(constants.GroupAddMemberFailure, errors)
}

// CreateGroupAddMemberSuccessAction .
func CreateGroupAddMemberSuccessAction() *Action {
	return createEmptyAction(constants.GroupAddMemberSuccess)
}

// CreateGroupCreateSuccessAction .
func CreateGroupCreateSuccessAction(
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

// CreateGroupLeaveFailureAction .
func CreateGroupLeaveFailureAction(errors []string) *Action {
	return createFailureAction(constants.GroupLeaveFailure, errors)
}

// CreateGroupLeaveSuccessAction .
func CreateGroupLeaveSuccessAction() *Action {
	return createEmptyAction(constants.GroupLeaveSuccess)
}

// CreateMessagingDeliverRequestAction .
func CreateMessagingDeliverRequestAction(message *db.Message) *Action {
	payload := make(map[string]interface{})
	payload["content"] = message.Content
	payload["date_created"] = message.DateCreated
	payload["id"] = message.ID
	payload["sender_id"] = message.SenderID

	return &Action{
		Payload: payload,
		Type:    constants.MessagingDeliverRequest,
	}
}

// CreateMessagingSendFailure .
func CreateMessagingSendFailure(errors []string) *Action {
	return createFailureAction(constants.MessagingSendFailure, errors)
}

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
