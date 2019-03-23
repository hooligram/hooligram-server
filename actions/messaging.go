package actions

// CreateMessagingSendFailure .
func CreateMessagingSendFailure(errors []string) *Action {
	return createFailureAction(MessagingSendFailure, errors)
}

// CreateMessagingSendSuccess .
func CreateMessagingSendSuccess(messageID int) *Action {
	payload := make(map[string]interface{})
	payload["message_id"] = messageID

	return &Action{
		Payload: payload,
		Type:    MessagingSendSuccess,
	}
}
