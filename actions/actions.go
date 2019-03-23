package actions

import (
	"github.com/hooligram/hooligram-server/db"
)

// Action .
type Action struct {
	Payload map[string]interface{} `json:"payload"`
	Type    string                 `json:"type"`
}

// CreateAuthorizationSignInFailure .
func CreateAuthorizationSignInFailure(errors []string) *Action {
	return createFailureAction(AuthorizationSignInFailure, errors)
}

// CreateGroupAddMemberFailure .
func CreateGroupAddMemberFailure(errors []string) *Action {
	return createFailureAction(GroupAddMemberFailure, errors)
}

// CreateGroupAddMemberSuccess .
func CreateGroupAddMemberSuccess() *Action {
	return createEmptyAction(GroupAddMemberSuccess)
}

// CreateGroupLeaveFailure .
func CreateGroupLeaveFailure(errors []string) *Action {
	return createFailureAction(GroupLeaveFailure, errors)
}

// CreateGroupLeaveSuccess .
func CreateGroupLeaveSuccess() *Action {
	return createEmptyAction(GroupLeaveSuccess)
}

// CreateMessagingDeliverRequest .
func CreateMessagingDeliverRequest(message *db.Message) *Action {
	payload := make(map[string]interface{})
	payload["content"] = message.Content
	payload["date_created"] = message.DateCreated
	payload["id"] = message.ID
	payload["sender_id"] = message.SenderID

	return &Action{
		Payload: payload,
		Type:    MessagingDeliverRequest,
	}
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
