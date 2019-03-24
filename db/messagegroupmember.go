package db

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

	return rows.Next()
}
