package db

////////////
// CREATE //
////////////

// CreateDirectMessage .
func CreateDirectMessage(groupID, memberAID, memberBID int) error {
	stmt, err := instance.Prepare(`
		INSERT INTO direct_message ( message_group_id, member_a_id, member_b_id )
		VALUES ( ?, ?, ? );
	`)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(groupID, memberAID, memberBID)
	if err != nil {
		return err
	}

	return nil
}

//////////
// READ //
//////////

// ReadDirectMessageGroupID .
func ReadDirectMessageGroupID(memberAID, memberBID int) (int, error) {
	rows, err := instance.Query(`
		SELECT message_group_id FROM direct_message
		WHERE
			member_a_id in (?, ?)
			AND
			member_b_id in (?, ?)
			AND
			member_a_id <> member_b_id
		LIMIT 1;
	`, memberAID, memberBID, memberAID, memberBID)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, nil
	}

	var messageGroupID int
	rows.Scan(&messageGroupID)
	return messageGroupID, nil
}

// ReadIsDirectMessage .
func ReadIsDirectMessage(groupID int) (bool, error) {
	rows, err := instance.Query(`
		SELECT COUNT(*) FROM direct_message WHERE message_group_id = ?;
	`, groupID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if !rows.Next() {
		return false, nil
	}

	var count int
	rows.Scan(&count)
	return count > 0, nil
}

////////////
// UPDATE //
////////////

////////////
// DELETE //
////////////
