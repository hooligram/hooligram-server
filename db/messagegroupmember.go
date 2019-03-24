package db

// FindAllMessageGroupMemberIDs .
func FindAllMessageGroupMemberIDs(messageGroupID int) ([]int, error) {
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

// IsClientInMessageGroup .
func IsClientInMessageGroup(clientID, messageGroupID int) bool {
	rows, err := instance.Query(`
		SELECT * FROM message_group_member
		WHERE message_group_id = ? AND member_id = ?;
	`, messageGroupID, clientID)

	if err != nil {
		return false
	}

	return rows.Next()
}
