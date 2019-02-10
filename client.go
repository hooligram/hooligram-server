package main

func (client *Client) writeEmptyAction(actionType string) {
	client.conn.WriteJSON(Action{
		Payload: map[string]interface{}{},
		Type:    actionType,
	})
}
