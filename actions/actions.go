package actions

const actionsTag = "actions"

// Action .
type Action struct {
	ID      string                 `json:"id"`
	Payload map[string]interface{} `json:"payload"`
	Type    string                 `json:"type"`
}

func constructEmptyAction(actionID, actionType string) *Action {
	payload := make(map[string]interface{})

	return &Action{
		ID:      actionID,
		Payload: payload,
		Type:    actionType,
	}
}

func constructFailureAction(actionID, actionType string, errors []string) *Action {
	payload := make(map[string]interface{})
	payload["errors"] = errors

	return &Action{
		ID:      actionID,
		Payload: payload,
		Type:    actionType,
	}
}
