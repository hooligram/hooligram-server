package db

// Message .
type Message struct {
	ID             int
	Content        string
	MessageGroupID int
	SenderID       int
	DateCreated    string
}
