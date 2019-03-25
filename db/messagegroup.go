package db

import (
	"errors"
)

// MessageGroup .
type MessageGroup struct {
	ID          int
	Name        string
	DateCreated string
}

// MemberIDs .
func (messageGroup *MessageGroup) MemberIDs() ([]int, error) {
	rows, err := instance.Query(`
		SELECT member_id FROM message_group_member WHERE message_group_id = ?;
	`, messageGroup.ID)
	if err != nil {
		return nil, err
	}

	memberIDs := []int{}

	for rows.Next() {
		var memberID int
		rows.Scan(&memberID)
		memberIDs = append(memberIDs, memberID)
	}

	return memberIDs, nil
}

////////////
// CREATE //
////////////

// CreateMessageGroup .
func CreateMessageGroup(groupName string, memberIDs []int) (*MessageGroup, error) {
	tx, err := instance.Begin()
	if err != nil {
		return nil, err
	}

	result, err := tx.Exec(
		`INSERT INTO message_group ( name ) VALUES ( ? );`,
		groupName,
	)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
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
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	rows, err := instance.Query(
		`SELECT date_created FROM message_group WHERE id = ?;`,
		groupID,
	)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		errorMsg := "message_group `%v` has been added to the database but "
		errorMsg += "an error occured when querying it"
		return nil, errors.New(errorMsg)
	}

	var dateCreated string
	rows.Scan(&dateCreated)

	messageGroup := &MessageGroup{
		ID:          int(groupID),
		Name:        groupName,
		DateCreated: dateCreated,
	}

	return messageGroup, nil
}

//////////
// READ //
//////////

// ReadMessageGroupByID .
func ReadMessageGroupByID(id int) (*MessageGroup, error) {
	rows, err := instance.Query(`
		SELECT id, name, date_created FROM message_group WHERE id = ?;
	`, id)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, nil
	}

	var name string
	var dateCreated string
	rows.Scan(&id, &name, &dateCreated)

	messageGroup := MessageGroup{
		ID:          id,
		Name:        name,
		DateCreated: dateCreated,
	}

	return &messageGroup, nil
}

////////////
// UPDATE //
////////////

////////////
// DELETE //
////////////
