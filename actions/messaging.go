package actions

// CreateMessagingDeliverSuccessFailure .
func CreateMessagingDeliverSuccessFailure(errors []string) *Action {
	return createFailureAction(MessagingDeliverSuccessFailure, errors)
}

// CreateMessagingDeliverSuccessSuccess .
func CreateMessagingDeliverSuccessSuccess() *Action {
	return createEmptyAction(MessagingDeliverSuccessSuccess)
}

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
