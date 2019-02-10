package main

func (client *Client) writeEmptyAction(actionType string) {
	writeEmptyAction(client.conn, actionType)
}

func (client *Client) writeJSON(action *Action) {
	client.conn.WriteJSON(*action)
}
