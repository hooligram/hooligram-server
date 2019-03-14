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
