package main

import (
	"database/sql"
	"errors"
	"log"
	"os"
)

func init() {
	dbUsername := os.Getenv("MYSQL_USERNAME")

	if dbUsername == "" {
		log.Println("[DB] MYSQL_USERNAME must be set.")
	}

	dbPassword := os.Getenv("MYSQL_PASSWORD")

	if dbPassword == "" {
		log.Println("[DB] MYSQL_PASSWORD must be set.")
	}

	dbName := os.Getenv("MYSQL_DB_NAME")

	if dbName == "" {
		log.Println("[DB] MYSQL_DB_NAME must be set.")
	}

	var err error
	db, err = sql.Open("mysql", dbUsername+":"+dbPassword+"@/"+dbName)

	if err != nil {
		log.Println("[DB] Error setting up MySQL DB connection.")
		return
	}

	err = db.Ping()

	if err != nil {
		log.Println("[DB] Can't connect to MySQL DB.")
		return
	}

	db.Exec("SET GLOBAL time_zone = '+00:00';")

	db.Exec(`
		CREATE TABLE IF NOT EXISTS client (
			id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			country_code VARCHAR (64) NOT NULL,
			phone_number VARCHAR (64) NOT NULL,
			verification_code VARCHAR (64),
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY (country_code, phone_number)
		);
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS message_group (
			id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			name VARCHAR (32) NOT NULL,
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id)
		);
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS message (
			id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			content TEXT NOT NULL,
			message_group_id INT UNSIGNED NOT NULL,
			sender_id INT UNSIGNED NOT NULL,
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			CONSTRAINT FOREIGN KEY (message_group_id) REFERENCES message_group (id)
				ON DELETE CASCADE
				ON UPDATE CASCADE,
			CONSTRAINT FOREIGN KEY (sender_id) REFERENCES client (id)
				ON DELETE CASCADE
				ON UPDATE CASCADE
		);
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS message_group_member (
			message_group_id INT UNSIGNED NOT NULL,
			member_id INT UNSIGNED NOT NULL,
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (message_group_id, member_id),
			CONSTRAINT FOREIGN KEY (message_group_id) REFERENCES message_group (id)
				ON DELETE CASCADE
				ON UPDATE CASCADE,
			CONSTRAINT FOREIGN KEY (member_id) REFERENCES client (id)
				ON DELETE CASCADE
				ON UPDATE CASCADE
		);
	`)

	db.Exec(`
		CREATE TABLE IF NOT EXISTS receipt (
			message_id INT UNSIGNED NOT NULL,
			recipient_id INT UNSIGNED NOT NULL,
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			date_delivered DATETIME,
			PRIMARY KEY (message_id, recipient_id),
			CONSTRAINT FOREIGN KEY (message_id) REFERENCES message (id)
				ON DELETE CASCADE
				ON UPDATE CASCADE,
			CONSTRAINT FOREIGN KEY (recipient_id) REFERENCES client (id)
				ON DELETE CASCADE
				ON UPDATE CASCADE
		);
	`)
}

func createMessage(content string, messageGroupID, senderID int) (*Message, error) {
	result, err := db.Exec(`
		INSERT INTO message (content, message_group_id, sender_id) VALUES (?, ?, ?);
	`, content, messageGroupID, senderID)

	if err != nil {
		return nil, errors.New("failed to create message")
	}

	id, _ := result.LastInsertId()
	rows, err := db.Query(`
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

func createReceipt(messageID, recipientID int) error {
	_, err := db.Query(`
		INSERT INTO receipt (message_id, recipient_id) VALUES (?, ?);
	`, messageID, recipientID)

	return err
}

func getOrCreateClient(countryCode, phoneNumber string) (*Client, error) {
	if countryCode != getDigits(countryCode) {
		return nil, errors.New("hey, countryCode should only contain digits")
	}

	if phoneNumber != getDigits(phoneNumber) {
		return nil, errors.New("hey, phoneNumber should only contain digits")
	}

	client, ok := findClient(countryCode, phoneNumber)

	if ok {
		return client, nil
	}

	_, err := db.Exec(`
		INSERT INTO client (country_code, phone_number) VALUES (?, ?)
	`, countryCode, phoneNumber)

	if err != nil {
		return nil, errors.New("i failed to create the client")
	}

	client, ok = findClient(countryCode, phoneNumber)

	if !ok {
		return nil, errors.New("i failed to find the created client")
	}

	return client, nil
}

func findAllMessageGroupMemberIDs(messageGroupID int) ([]int, error) {
	rows, err := db.Query(`
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

func findAllVerifiedClients() []*Client {
	rows, err := db.Query(`
		SELECT *
		FROM client
		WHERE verification_code IS NOT NULL;
	`)
	clients := []*Client{}

	if err != nil {
		log.Println("[DB] Failed to find all verified clients.")
		return clients
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

	return clients
}

func findClient(countryCode, phoneNumber string) (*Client, bool) {
	rows, err := db.Query(`
		SELECT *
		FROM client
		WHERE
			country_code = ?
			AND
			phone_number = ?
	`, countryCode, phoneNumber)

	if err != nil {
		log.Println("[DB] i failed to find the client")
		log.Println("[DB]", err.Error())
		return nil, false
	}

	if !rows.Next() {
		return nil, false
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

	return client, true
}

func findUndeliveredMessages(recipientID int) ([]*Message, error) {
	rows, err := db.Query(`
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

func findVerifiedClient(countryCode, phoneNumber, verificationCode string) (*Client, bool) {
	rows, err := db.Query(`
		SELECT *
		FROM client
		WHERE
			country_code = ? AND phone_number = ? AND verification_code = ?;
	`, countryCode, phoneNumber, verificationCode)

	if err != nil {
		log.Println("[DB] Find client failed.")
		return nil, false
	}

	if !rows.Next() {
		log.Println("[DB] Couldn't find such client.")
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

func isClientInMessageGroup(clientID, messageGroupID int) bool {
	rows, err := db.Query(`
		SELECT * FROM message_group_member
		WHERE message_group_id = ? AND member_id = ?;
	`, clientID, messageGroupID)

	if err != nil {
		return false
	}

	return rows.Next()
}

func updateClientVerificationCode(client *Client, verificationCode string) error {
	_, err := db.Exec(`
		UPDATE client SET verification_code = ? WHERE country_code = ? AND phone_number = ?;
	`, verificationCode, client.CountryCode, client.PhoneNumber)

	if err != nil {
		return err
	}

	return nil
}

func updateReceiptDateDelivered(messageID, recipientID int) error {
	_, err := db.Exec(`
		UPDATE receipt SET date_delivered = CURRENT_TIMESTAMP
		WHERE message_id = ? AND recipient_id = ?;
	`, messageID, recipientID)
	return err
}
