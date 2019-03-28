package db

import (
	"database/sql"
	"flag"
	"os"
	"regexp"

	"github.com/hooligram/hooligram-server/utils"
)

const dbTag = "db"

var (
	dbName     = os.Getenv("MYSQL_DB_NAME")
	dbPassword = os.Getenv("MYSQL_PASSWORD")
	dbUsername = os.Getenv("MYSQL_USERNAME")
	instance   *sql.DB
)

func init() {
	if flag.Lookup("test.v") != nil {
		return
	}

	if dbUsername == "" {
		utils.LogFatal(dbTag, "MYSQL_USERNAME not set")
	}

	if dbPassword == "" {
		utils.LogFatal(dbTag, "MYSQL_PASSWORD not set")
	}

	if dbName == "" {
		utils.LogFatal(dbTag, "MYSQL_DB_NAME not set")
	}

	var err error
	instance, err = sql.Open("mysql", dbUsername+":"+dbPassword+"@/"+dbName)
	if err != nil {
		utils.LogFatal(dbTag, "error opening mysql connection. "+err.Error())
	}

	err = instance.Ping()
	if err != nil {
		utils.LogFatal(dbTag, "error pinging mysql. "+err.Error())
	}

	_, err = instance.Exec("SET GLOBAL time_zone = '+00:00';")
	if err != nil {
		utils.LogFatal(dbTag, "error setting global time zone. "+err.Error())
	}

	_, err = instance.Exec(`
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
	if err != nil {
		utils.LogFatal(dbTag, "error creating client table. "+err.Error())
	}

	_, err = instance.Exec(`
		CREATE TABLE IF NOT EXISTS message_group (
			id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			name VARCHAR ( 32 ) NOT NULL,
			date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY ( id )
		);
	`)
	if err != nil {
		utils.LogFatal(dbTag, "error creating message_group table. "+err.Error())
	}

	_, err = instance.Exec(`
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
	if err != nil {
		utils.LogFatal(dbTag, "error creating message table")
	}

	_, err = instance.Exec(`
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
	if err != nil {
		utils.LogFatal(dbTag, "error creating message_group_member table")
	}

	_, err = instance.Exec(`
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
	if err != nil {
		utils.LogFatal(dbTag, "error creating receipt table")
	}
}

// ReadClientMessageGroupIDs .
func ReadClientMessageGroupIDs(clientID int) ([]int, error) {
	rows, err := instance.Query(`
		SELECT message_group.id
		FROM message_group_member
			JOIN message_group
			ON message_group_member.message_group_id = message_group.id
		WHERE message_group_member.member_id = ?;
	`, clientID)
	if err != nil {
		return nil, err
	}

	groupIDs := []int{}

	for rows.Next() {
		var groupID int
		rows.Scan(&groupID)
		groupIDs = append(groupIDs, groupID)
	}

	return groupIDs, nil
}

// ReadUndeliveredMessages .
func ReadUndeliveredMessages(recipientID int) ([]*Message, error) {
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

func getDigits(s string) string {
	re := regexp.MustCompile("[^0-9]")
	return re.ReplaceAllString(s, "")
}
