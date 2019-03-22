package db

import (
	"errors"
)

// Client .
type Client struct {
	ID               int
	CountryCode      string
	PhoneNumber      string
	VerificationCode string
	DateCreated      string
}

// CreateClient .
func CreateClient(countryCode, phoneNumber string) (*Client, error) {
	if countryCode != getDigits(countryCode) {
		return nil, errors.New("country code should only contain digits")
	}

	if phoneNumber != getDigits(phoneNumber) {
		return nil, errors.New("phone number should only contain digits")
	}

	client, err := ReadClientByUniqueKey(countryCode, phoneNumber)
	if err != nil {
		return nil, err
	}

	if client != nil {
		return client, nil
	}

	result, err := instance.Exec(`
		INSERT INTO client ( country_code, phone_number ) VALUES ( ?, ? );
	`, countryCode, phoneNumber)
	if err != nil {
		return nil, err
	}

	clientID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	client, err = ReadClientByID(int(clientID))
	if err != nil {
		return nil, err
	}

	return client, nil
}

// ReadClientByID .
func ReadClientByID(id int) (*Client, error) {
	rows, err := instance.Query(`
		SELECT country_code, phone_number, verification_code, date_created FROM client WHERE id = ?;
	`, id)

	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, nil
	}

	var countryCode string
	var phoneNumber string
	var verificationCode string
	var dateCreated string
	rows.Scan(&countryCode, &phoneNumber, &verificationCode, &dateCreated)

	client := &Client{
		ID:               id,
		CountryCode:      countryCode,
		PhoneNumber:      phoneNumber,
		VerificationCode: verificationCode,
		DateCreated:      dateCreated,
	}

	return client, nil
}

// ReadClientByUniqueKey .
func ReadClientByUniqueKey(countryCode, phoneNumber string) (*Client, error) {
	rows, err := instance.Query(`
		SELECT id, verification_code, date_created
		FROM client
		WHERE country_code = ? AND phone_number = ?;
	`, countryCode, phoneNumber)
	if err != nil {
		return nil, err
	}

	if !rows.Next() {
		return nil, nil
	}

	var id int
	var verificationCode string
	var dateCreated string
	rows.Scan(&id, &verificationCode, &dateCreated)

	client := &Client{
		ID:               id,
		CountryCode:      countryCode,
		PhoneNumber:      phoneNumber,
		VerificationCode: verificationCode,
		DateCreated:      dateCreated,
	}

	return client, nil
}

// ReadClientVerificationCode .
func ReadClientVerificationCode(clientID int) (string, error) {
	rows, err := instance.Query(`
	SELECT verification_code FROM client WHERE id = ?;
	`, clientID)
	if err != nil {
		return "", err
	}

	if !rows.Next() {
		return "", nil
	}

	var verificationCode string
	rows.Scan(&verificationCode)
	return verificationCode, nil
}

// UpdateClientVerificationCode .
func UpdateClientVerificationCode(clientID int, verificationCode string) error {
	_, err := instance.Exec(`
		UPDATE client SET verification_code = ? WHERE id = ?;
	`, verificationCode, clientID)
	return err
}
