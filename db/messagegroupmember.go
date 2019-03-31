package db

////////////
// CREATE //
////////////

// CreateMessageGroupMembers .
func CreateMessageGroupMembers(messageGroupID int, memberIDs []int) error {
	tx, err := instance.Begin()
	if err != nil {
		return err
	}

	insert, err := tx.Prepare(`
		INSERT INTO message_group_member ( message_group_id, member_id )
		VALUES ( ?, ? );
	`)
	if err != nil {
		return err
	}

	for _, memberID := range memberIDs {
		_, err := insert.Exec(messageGroupID, memberID)
		if err != nil {
			return tx.Rollback()
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

//////////
// READ //
//////////

// ReadMessageGroupMemberIDs .
func ReadMessageGroupMemberIDs(messageGroupID int) ([]int, error) {
	rows, err := instance.Query(`
		SELECT member_id
		FROM message_group_member
		WHERE message_group_id = ?;
	`, messageGroupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memberIDs []int

	for rows.Next() {
		var id int
		rows.Scan(&id)
		memberIDs = append(memberIDs, id)
	}

	return memberIDs, nil
}

// ReadIsClientInMessageGroup .
func ReadIsClientInMessageGroup(clientID, messageGroupID int) bool {
	rows, err := instance.Query(`
		SELECT * FROM message_group_member
		WHERE message_group_id = ? AND member_id = ?;
	`, messageGroupID, clientID)
	if err != nil {
		return false
	}
	defer rows.Close()

	return rows.Next()
}

////////////
// UPDATE //
////////////

////////////
// DELETE //
////////////

// DeleteMessageGroupMembers .
func DeleteMessageGroupMembers(groupID int, memberIDs []int) error {
	tx, err := instance.Begin()
	if err != nil {
		return err
	}

	delete, err := tx.Prepare(`
		DELETE FROM message_group_member WHERE message_group_id = ? AND member_id = ?;
	`)
	if err != nil {
		return err
	}

	for _, memberID := range memberIDs {
		_, err := delete.Exec(groupID, memberID)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}
