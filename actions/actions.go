package actions

const actionsTag = "actions"

// Action .
type Action struct {
	Payload map[string]interface{} `json:"payload"`
	Type    string                 `json:"type"`
}

func constructEmptyAction(actionType string) *Action {
	payload := make(map[string]interface{})

	return &Action{
		Payload: payload,
		Type:    actionType,
	}
}

func constructFailureAction(actionType string, errors []string) *Action {
	payload := make(map[string]interface{})
	payload["errors"] = errors

	return &Action{
		Payload: payload,
		Type:    actionType,
	}
}
