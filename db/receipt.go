package db

////////////
// CREATE //
////////////

// CreateReceipt .
func CreateReceipt(messageID, recipientID int) error {
	_, err := instance.Query(`
		INSERT INTO receipt ( message_id, recipient_id ) VALUES ( ?, ? );
	`, messageID, recipientID)

	return err
}

//////////
// READ //
//////////

////////////
// UPDATE //
////////////

// UpdateReceiptDateDelivered .
func UpdateReceiptDateDelivered(messageID, recipientID int) error {
	_, err := instance.Exec(`
		UPDATE receipt SET date_delivered = CURRENT_TIMESTAMP
		WHERE message_id = ? AND recipient_id = ?;
	`, messageID, recipientID)
	return err
}
