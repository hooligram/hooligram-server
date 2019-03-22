package db

// Client .
type Client struct {
	ID               int
	CountryCode      string
	PhoneNumber      string
	VerificationCode string
	DateCreated      string
}

// ReadVerificationCode .
func ReadVerificationCode(clientID int) (string, error) {
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

// UpdateVerificationCode .
func UpdateVerificationCode(clientID int, verificationCode string) error {
	_, err := instance.Exec(`
		UPDATE client SET verification_code = ? WHERE id = ?;
	`, verificationCode, clientID)
	return err
}
