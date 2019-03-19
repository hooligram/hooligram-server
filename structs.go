package main

import (
	"github.com/gorilla/websocket"
)

// Action .
type Action struct {
	Payload map[string]interface{} `json:"payload"`
	Type    string                 `json:"type"`
}

// Client .
type Client struct {
	ID               int
	CountryCode      string
	PhoneNumber      string
	VerificationCode string
	DateCreated      string

	IsSignedIn bool
	conn       *websocket.Conn
}

// Message .
type Message struct {
	ID             int
	Content        string
	MessageGroupID int
	SenderID       int
	DateCreated    string
}

// MessageDelivery .
type MessageDelivery struct {
	Message      *Message
	RecipientIDs []int
}

// MessageGroup .
type MessageGroup struct {
	ID int64
	DateCreated string
	MemberIDs []int
	Name string
}
