package main

import (
	"database/sql"
	"log"
	"os"
)

func init() {
	dbUsername := os.Getenv("MYSQL_USERNAME")

	if dbUsername == "" {
		log.Fatal("[DB] MYSQL_USERNAME must be set. Exiting...")
	}

	dbPassword := os.Getenv("MYSQL_PASSWORD")

	if dbPassword == "" {
		log.Fatal("[DB] MYSQL_PASSWORD must be set. Exiting...")
	}

	dbName := os.Getenv("MYSQL_DB_NAME")

	if dbName == "" {
		log.Fatal("[DB] MYSQL_DB_NAME must be set. Exiting...")
	}

	var err error
	db, err = sql.Open("mysql", dbUsername+":"+dbPassword+"@/"+dbName)

	if err != nil {
		log.Fatal("[DB] Error setting up MySQL DB connection. Exiting...")
	}

	err = db.Ping()

	if err != nil {
		log.Fatal("[DB] Can't connect to MySQL DB. Exiting...")
	}

	db.Exec(`
		CREATE TABLE IF NOT EXISTS client (
			id INT UNSIGNED NOT NULL AUTO_INCREMENT,
			country_code VARCHAR (50) NOT NULL,
			phone_number VARCHAR (50) NOT NULL,
			verification_code VARCHAR (50),
			PRIMARY KEY (id)
		)
	`)
}

func createClient(countryCode, phoneNumber string) bool {
	_, err := db.Exec(`
		INSERT INTO client (country_code, phone_number) VALUES (?, ?)
	`, countryCode, phoneNumber)

	if err != nil {
		log.Println("[DB] Create client failed.")
		return false
	}

	return true
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
		rows.Scan(&id, &countryCode, &phoneNumber, &verificationCode)
		client := Client{
			ID:               id,
			CountryCode:      countryCode,
			PhoneNumber:      phoneNumber,
			VerificationCode: verificationCode,
		}
		clients = append(clients, &client)
	}

	return clients
}

func findClient(countryCode, phoneNumber string) bool {
	rows, err := db.Query(`
		SELECT *
		FROM client
		WHERE
			country_code = ?
			AND
			phone_number = ?
	`, countryCode, phoneNumber)

	if err != nil {
		log.Println("[DB] Find client failed.")
		return false
	}

	if !rows.Next() {
		return false
	}

	return true
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
	rows.Scan(&id)
	client := Client{
		ID:               id,
		CountryCode:      countryCode,
		PhoneNumber:      phoneNumber,
		VerificationCode: verificationCode,
	}

	return &client, true
}

func updateClientVerificationCode(client *Client, verificationCode string) bool {
	_, err := db.Exec(`
		UPDATE client SET verification_code = ? WHERE country_code = ? AND phone_number = ?;
	`, verificationCode, client.CountryCode, client.PhoneNumber)

	if err != nil {
		log.Println("[DB] Failed to update client's verification code record.")
		return false
	}

	return true
}
