package structs

// IsVerified .
func (client *Client) IsVerified() bool {
	return client.VerificationCode != ""
}

// WriteEmptyAction .
func (client *Client) WriteEmptyAction(actionType string) {
	// utils.writeEmptyAction(client.conn, actionType)
	client.Conn.WriteJSON(Action{
		map[string]interface{}{},
		actionType,
	})
}

// WriteFailure .
func (client *Client) WriteFailure(actionType string, errors []string) {
	// utils.writeFailure(client.conn, actionType, errors)
	client.Conn.WriteJSON(Action{
		map[string]interface{}{
			"errors": errors,
		},
		actionType,
	})
}

// WriteJSON .
func (client *Client) WriteJSON(action *Action) {
	client.Conn.WriteJSON(*action)
}
