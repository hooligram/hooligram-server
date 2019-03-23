package clients

import (
	"github.com/gorilla/websocket"
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/db"
)

// Client .
type Client struct {
	Conn      *websocket.Conn
	SessionID string

	id         int
	isSignedIn bool
}

// GetID .
func (client *Client) GetID() int {
	return client.id
}

// GetVerificationCode .
func (client *Client) GetVerificationCode() (string, error) {
	if client.id < 1 {
		return "", nil
	}

	clientRow, err := db.ReadClientByID(client.id)
	if err != nil {
		return "", err
	}

	return clientRow.VerificationCode, nil
}

// IsVerified .
func (client *Client) IsVerified() (bool, error) {
	verificationCode, err := db.ReadClientVerificationCode(client.id)
	if err != nil {
		return false, err
	}

	return verificationCode != "", nil
}

// IsSignedIn .
func (client *Client) IsSignedIn() bool {
	return client.isSignedIn
}

// Register .
func (client *Client) Register(countryCode, phoneNumber string) (bool, error) {
	clientRow, err := db.ReadClientByUniqueKey(countryCode, phoneNumber)
	if err != nil {
		return false, err
	}

	if clientRow == nil {
		clientRow, err = db.CreateClient(countryCode, phoneNumber)
		if err != nil {
			return false, err
		}
	} else {
		err := db.UpdateClientVerificationCode(clientRow.ID, "")
		if err != nil {
			return false, err
		}
	}

	client.id = clientRow.ID
	return true, nil
}

// SignIn .
func (client *Client) SignIn(
	countryCode string,
	phoneNumber string,
	verificationCode string,
) (bool, error) {
	row, err := db.ReadClientByUniqueKey(countryCode, phoneNumber)
	if err != nil {
		return false, err
	}

	if row == nil {
		return false, nil
	}

	if row.VerificationCode != verificationCode {
		return false, nil
	}

	client.id = row.ID
	client.isSignedIn = true
	return true, nil
}

// WriteEmptyAction .
func (client *Client) WriteEmptyAction(actionType string) {
	client.WriteJSON(&actions.Action{
		Payload: map[string]interface{}{},
		Type:    actionType,
	})
}

// WriteFailure .
func (client *Client) WriteFailure(actionType string, errors []string) {
	client.WriteJSON(&actions.Action{
		Payload: map[string]interface{}{
			"errors": errors,
		},
		Type: actionType,
	})
}

// WriteJSON .
func (client *Client) WriteJSON(action *actions.Action) error {
	return client.Conn.WriteJSON(*action)
}
