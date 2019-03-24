package db

import "errors"

// MessageGroup .
type MessageGroup struct {
	ID          int64
	Name        string
	DateCreated string

	MemberIDs []int
}

////////////
// CREATE //
////////////

// CreateMessageGroup .
func CreateMessageGroup(groupName string, memberIDs []int) (*MessageGroup, error) {
	tx, err := instance.Begin()
	if err != nil {
		// utils.LogInfo(dbTag, "transaction error. "+err.Error())
		return nil, err
	}

	result, err := tx.Exec(
		`INSERT INTO message_group ( name ) VALUES ( ? );`,
		groupName,
	)
	if err != nil {
		tx.Rollback()
		// utils.LogInfo(dbTag, "error creating message group. "+err.Error())
		return nil, err
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		// utils.LogInfo(dbTag, "error creating message group. "+err.Error())
		return nil, err
	}

	for _, memberID := range memberIDs {
		result, err = tx.Exec(
			`INSERT INTO message_group_member ( message_group_id, member_id )
			VALUES ( ?, ? );`,
			groupID,
			memberID,
		)
		if err != nil {
			tx.Rollback()
			// utils.LogInfo(
			// 	dbTag,
			// 	fmt.Sprintf("failed to create message group %v in instance. %v", groupName, err.Error()),
			// )
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		// utils.LogInfo(dbTag, "error committing transaction. "+err.Error())
		return nil, err
	}

	rows, err := instance.Query(
		`SELECT date_created FROM message_group WHERE id = ?;`,
		groupID,
	)
	if err != nil {
		// utils.LogInfo(dbTag, "error retrieving message group. "+err.Error())
		return nil, err
	}

	if !rows.Next() {
		errorMsg := "message_group `%v` has been added to the database but "
		errorMsg += "an error occured when querying it"
		// utils.LogInfo(dbTag, fmt.Sprintf(errorMsg, groupName))
		return nil, errors.New(errorMsg)
	}

	var dateCreated string
	rows.Scan(&dateCreated)

	messageGroup := &MessageGroup{
		ID:          groupID,
		Name:        groupName,
		DateCreated: dateCreated,

		MemberIDs: memberIDs,
	}

	return messageGroup, nil
}

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
