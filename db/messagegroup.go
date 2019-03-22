package db

// MessageGroup .
type MessageGroup struct {
	ID          int64
	Name        string
	DateCreated string

	MemberIDs []int
}
