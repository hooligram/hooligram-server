package main

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]*Client)
var db *sql.DB
var httpClient = &http.Client{}
var twilioAPIKey string
var upgrader = websocket.Upgrader{}
