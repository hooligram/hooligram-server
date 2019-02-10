package main

func (client *Client) writeEmptyAction(actionType string) {
	writeEmptyAction(client.conn, actionType)
}

func (client *Client) writeFailure(actionType string, errors []string) {
	writeFailure(client.conn, actionType, errors)
}

func (client *Client) writeJSON(action *Action) {
	client.conn.WriteJSON(*action)
}
