package db

////////////
// CREATE //
////////////

// CreateReceipt .
func CreateReceipt(messageID, recipientID int) error {
	rows, err := instance.Query(`
		INSERT INTO receipt ( message_id, recipient_id ) VALUES ( ?, ? );
	`, messageID, recipientID)
	defer rows.Close()

	return err
}

//////////
// READ //
//////////

////////////
// UPDATE //
////////////

// UpdateReceiptDateDelivered .
func UpdateReceiptDateDelivered(messageID, recipientID int) (bool, error) {
	result, err := instance.Exec(`
		UPDATE receipt SET date_delivered = CURRENT_TIMESTAMP
		WHERE message_id = ? AND recipient_id = ?;
	`, messageID, recipientID)
	if err != nil {
		return false, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}

	return rows > 0, nil
}
