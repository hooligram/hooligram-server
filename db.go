package main

import (
	"database/sql"
	"log"
	"os"
)

func init() {
	log.Println("[DB] Initializing DB...")

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

	log.Println("[DB] Creating DB tables...")
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

	log.Printf(
		"[DB] Created client (country_code: %s, phone_number: %s)\n",
		countryCode,
		phoneNumber,
	)
	return true
}

func findClient(countryCode, phoneNumber string) bool {
	log.Println("[DB] Finding client...")

	rows, err := db.Query(`
		SELECT *
		FROM client
		WHERE
			country_code = ?
			AND
			phone_number = ?
	`, countryCode, phoneNumber)

	if err != nil {
		log.Println("[DB] Failed to find client.")
		return false
	}

	if !rows.Next() {
		log.Println("[DB] Client not found.")
		return false
	}

	log.Println("[DB] Client found.")
	return true
}

func findVerifiedClient(countryCode, phoneNumber, verificationCode string) bool {
	rows, _ := db.Query(`
		SELECT *
		FROM client
		WHERE
			country_code = ? AND phone_number = ? AND verification_code = ?;
	`, countryCode, phoneNumber, verificationCode)

	if !rows.Next() {
		return false
	}

	return true
}
