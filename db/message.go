package db

import (
	"errors"

	"github.com/hooligram/hooligram-server/utils"
)

// Message .
type Message struct {
	ID             int
	Content        string
	MessageGroupID int
	SenderID       int
	DateCreated    string
}

// SenderSID .
func (message *Message) SenderSID() string {
	clientRow, err := ReadClientByID(message.SenderID)
	if err != nil {
		utils.LogInfo(dbTag, "error reading client by id. "+err.Error())
		return ""
	}

	return clientRow.CountryCode + "." + clientRow.PhoneNumber
}

////////////
// CREATE //
////////////

// CreateMessage .
func CreateMessage(content string, messageGroupID, senderID int) (*Message, error) {
	result, err := instance.Exec(`
		INSERT INTO message ( content, message_group_id, sender_id ) VALUES ( ?, ?, ? );
	`, content, messageGroupID, senderID)

	if err != nil {
		return nil, errors.New("failed to create message")
	}

	id, _ := result.LastInsertId()
	rows, err := instance.Query(`
		SELECT date_created FROM message WHERE id = ?;
	`, id)

	if err != nil {
		return nil, errors.New("failed to find message")
	}

	if !rows.Next() {
		return nil, errors.New("can't find message")
	}

	var dateCreated string
	rows.Scan(&dateCreated)

	return &Message{
		ID:             int(id),
		Content:        content,
		MessageGroupID: messageGroupID,
		SenderID:       senderID,
		DateCreated:    dateCreated,
	}, nil
}
