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

// MemberSIDs .
func (messageGroup *MessageGroup) MemberSIDs() ([]string, error) {
	rows, err := instance.Query(`
		SELECT client.country_code, client.phone_number
		FROM message_group_member
			JOIN client ON message_group_member.member_id = client.id
		WHERE message_group_id = ?;
	`, messageGroup.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	memberSIDs := []string{}

	for rows.Next() {
		var countryCode string
		var phoneNumber string
		rows.Scan(&countryCode, &phoneNumber)
		sid := countryCode + "." + phoneNumber
		memberSIDs = append(memberSIDs, sid)
	}

	return memberSIDs, nil
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
	defer rows.Close()

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
	defer rows.Close()

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

// DeleteMessageGroup .
func DeleteMessageGroup(id int) error {
	_, err := instance.Exec("DELETE FROM message_group WHERE id = ?;", id)
	return err
}
