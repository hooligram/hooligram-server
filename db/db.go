package db

import (
	"database/sql"
	"errors"
	"os"
	"regexp"
)

const dbTag = "db"

var instance *sql.DB

func init() {
	dbUsername := os.Getenv("MYSQL_USERNAME")
	if dbUsername == "" {
		// utils.LogInfo(dbTag, "MYSQL_USERNAME not set")
	}

	dbPassword := os.Getenv("MYSQL_PASSWORD")
	if dbPassword == "" {
		// utils.LogInfo(dbTag, "MYSQL_PASSWORD not set")
	}

	dbName := os.Getenv("MYSQL_DB_NAME")
	if dbName == "" {
		// utils.LogInfo(dbTag, "MYSQL_DB_NAME not set")
	}

	var err error
	instance, err = sql.Open("mysql", dbUsername+":"+dbPassword+"@/"+dbName)
	if err != nil {
		// utils.LogInfo(dbTag, "mysql connection setup error. "+err.Error())
		return
	}

	err = instance.Ping()
	if err != nil {
		// utils.LogInfo(dbTag, "mysql connection error. "+err.Error())
		return
	}

	instance.Exec("SET GLOBAL time_zone = '+00:00';")

	instance.Exec(`
		CREATE TABLE IF NOT EXISTS client (
			id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			country_code VARCHAR ( 64 ) NOT NULL,
			phone_number VARCHAR ( 64 ) NOT NULL,
			verification_code VARCHAR ( 64 ),
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY ( id ),
			UNIQUE KEY ( country_code, phone_number )
		);
	`)

	instance.Exec(`
		CREATE TABLE IF NOT EXISTS message_group (
			id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			name VARCHAR ( 32 ) NOT NULL,
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY ( id )
		);
	`)

	instance.Exec(`
		CREATE TABLE IF NOT EXISTS message (
			id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			content TEXT NOT NULL,
			message_group_id INT UNSIGNED NOT NULL,
			sender_id INT UNSIGNED NOT NULL,
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY ( id ),
			CONSTRAINT FOREIGN KEY ( message_group_id ) REFERENCES message_group ( id )
				ON DELETE CASCADE
				ON UPDATE CASCADE,
			CONSTRAINT FOREIGN KEY ( sender_id ) REFERENCES client ( id )
				ON DELETE CASCADE
				ON UPDATE CASCADE
		);
	`)

	instance.Exec(`
		CREATE TABLE IF NOT EXISTS message_group_member (
			message_group_id INT UNSIGNED NOT NULL,
			member_id INT UNSIGNED NOT NULL,
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY ( message_group_id, member_id ),
			CONSTRAINT FOREIGN KEY ( message_group_id ) REFERENCES message_group ( id )
				ON DELETE CASCADE
				ON UPDATE CASCADE,
			CONSTRAINT FOREIGN KEY ( member_id ) REFERENCES client ( id )
				ON DELETE CASCADE
				ON UPDATE CASCADE
		);
	`)

	instance.Exec(`
		CREATE TABLE IF NOT EXISTS receipt (
			message_id INT UNSIGNED NOT NULL,
			recipient_id INT UNSIGNED NOT NULL,
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			date_delivered DATETIME,
			PRIMARY KEY ( message_id, recipient_id ),
			CONSTRAINT FOREIGN KEY ( message_id ) REFERENCES message ( id )
				ON DELETE CASCADE
				ON UPDATE CASCADE,
			CONSTRAINT FOREIGN KEY ( recipient_id ) REFERENCES client ( id )
				ON DELETE CASCADE
				ON UPDATE CASCADE
		);
	`)
}

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

// CreateReceipt .
func CreateReceipt(messageID, recipientID int) error {
	_, err := instance.Query(`
		INSERT INTO receipt ( message_id, recipient_id ) VALUES ( ?, ? );
	`, messageID, recipientID)

	return err
}

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

// GetOrCreateClient .
func GetOrCreateClient(countryCode, phoneNumber string) (*Client, error) {
	if countryCode != getDigits(countryCode) {
		return nil, errors.New("country code should only contain digits")
	}

	if phoneNumber != getDigits(phoneNumber) {
		return nil, errors.New("phone number should only contain digits")
	}

	client, err := FindClient(countryCode, phoneNumber)
	if err != nil {
		return nil, err
	}

	if client != nil {
		return client, nil
	}

	_, err = instance.Exec(`
		INSERT INTO client ( country_code, phone_number ) VALUES ( ?, ? );
	`, countryCode, phoneNumber)
	if err != nil {
		return nil, err
	}

	client, err = FindClient(countryCode, phoneNumber)
	if err != nil {
		return nil, err
	}

	return client, nil
}

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

// FindAllVerifiedClients .
func FindAllVerifiedClients() ([]*Client, error) {
	rows, err := instance.Query(`
		SELECT *
		FROM client
		WHERE verification_code IS NOT NULL;
	`)
	clients := []*Client{}

	if err != nil {
		// utils.LogInfo(dbTag, "failed to find verified clients. "+err.Error())
		return clients, err
	}

	for rows.Next() {
		var id int
		var countryCode string
		var phoneNumber string
		var verificationCode string
		var dateCreated string
		rows.Scan(&id, &countryCode, &phoneNumber, &verificationCode, &dateCreated)
		client := Client{
			ID:               id,
			CountryCode:      countryCode,
			PhoneNumber:      phoneNumber,
			VerificationCode: verificationCode,
			DateCreated:      dateCreated,
		}
		clients = append(clients, &client)
	}

	return clients, nil
}

// FindClient .
func FindClient(countryCode, phoneNumber string) (*Client, error) {
	rows, err := instance.Query(`
		SELECT *
		FROM client
		WHERE
			country_code = ?
			AND
			phone_number = ?
	`, countryCode, phoneNumber)

	if err != nil {
		// utils.LogInfo(dbTag, "failed to find client. "+err.Error())
		return nil, err
	}

	if !rows.Next() {
		return nil, nil
	}

	var id int
	var verificationCode string
	var dateCreated string
	rows.Scan(&id, nil, nil, &verificationCode, &dateCreated)

	client := &Client{
		ID:               id,
		CountryCode:      countryCode,
		PhoneNumber:      phoneNumber,
		VerificationCode: verificationCode,
		DateCreated:      dateCreated,
	}

	return client, nil
}

// FindUndeliveredMessages .
func FindUndeliveredMessages(recipientID int) ([]*Message, error) {
	rows, err := instance.Query(`
		SELECT message.*
		FROM receipt JOIN message ON receipt.message_id = message.id
		WHERE receipt.recipient_id = ? AND receipt.date_delivered IS NULL;
	`, recipientID)
	if err != nil {
		return nil, err
	}

	var messages []*Message

	for rows.Next() {
		var id int
		var content string
		var messageGroupID int
		var senderID int
		var dateCreated string
		rows.Scan(&id, &content, &messageGroupID, &senderID, &dateCreated)
		messages = append(messages, &Message{
			ID:             id,
			Content:        content,
			MessageGroupID: messageGroupID,
			SenderID:       senderID,
			DateCreated:    dateCreated,
		})
	}

	return messages, nil
}

// FindVerifiedClient .
func FindVerifiedClient(countryCode, phoneNumber, verificationCode string) (*Client, bool) {
	rows, err := instance.Query(`
		SELECT *
		FROM client
		WHERE
			country_code = ? AND phone_number = ? AND verification_code = ?;
	`, countryCode, phoneNumber, verificationCode)

	if err != nil {
		// utils.LogInfo(dbTag, "failed to find client. "+err.Error())
		return nil, false
	}

	if !rows.Next() {
		return nil, false
	}

	var id int
	var dateCreated string
	rows.Scan(&id, nil, nil, nil, &dateCreated)
	client := Client{
		ID:               id,
		CountryCode:      countryCode,
		PhoneNumber:      phoneNumber,
		VerificationCode: verificationCode,
		DateCreated:      dateCreated,
	}

	return &client, true
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

// UpdateClientVerificationCode .
func UpdateClientVerificationCode(client *Client, verificationCode string) error {
	_, err := instance.Exec(`
		UPDATE client SET verification_code = ? WHERE country_code = ? AND phone_number = ?;
	`, verificationCode, client.CountryCode, client.PhoneNumber)

	if err != nil {
		return err
	}

	return nil
}

// UpdateReceiptDateDelivered .
func UpdateReceiptDateDelivered(messageID, recipientID int) error {
	_, err := instance.Exec(`
		UPDATE receipt SET date_delivered = CURRENT_TIMESTAMP
		WHERE message_id = ? AND recipient_id = ?;
	`, messageID, recipientID)
	return err
}

func getDigits(s string) string {
	re := regexp.MustCompile("[^0-9]")
	return re.ReplaceAllString(s, "")
}
