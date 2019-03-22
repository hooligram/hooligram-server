package clients

import (
	"github.com/gorilla/websocket"
	"github.com/hooligram/hooligram-server/actions"
	"github.com/hooligram/hooligram-server/db"
)

// Client .
type Client struct {
	db.Client

	Conn       *websocket.Conn
	IsSignedIn bool
	SessionID  string
}

// IsVerified .
func (client *Client) IsVerified() (bool, error) {
	verificationCode, err := db.ReadVerificationCode(client.ID)
	if err != nil {
		return false, err
	}

	return verificationCode != "", nil
}

// SignIn .
func (client *Client) SignIn(
	countryCode string,
	phoneNumber string,
	verificationCode string,
) {
	client.CountryCode = countryCode
	client.PhoneNumber = phoneNumber
	client.VerificationCode = verificationCode

	client.IsSignedIn = true
}

// WriteEmptyAction .
func (client *Client) WriteEmptyAction(actionType string) {
	// utils.writeEmptyAction(client.conn, actionType)
	client.Conn.WriteJSON(actions.Action{
		map[string]interface{}{},
		actionType,
	})
}

// WriteFailure .
func (client *Client) WriteFailure(actionType string, errors []string) {
	// utils.writeFailure(client.conn, actionType, errors)
	client.Conn.WriteJSON(actions.Action{
		map[string]interface{}{
			"errors": errors,
		},
		actionType,
	})
}

// WriteJSON .
func (client *Client) WriteJSON(action *actions.Action) {
	client.Conn.WriteJSON(*action)
}
