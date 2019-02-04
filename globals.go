package main

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/websocket"
)

var broadcastChan = make(chan *Action)
var clients = make(map[*websocket.Conn]*Client)
var db *sql.DB
var httpClient = &http.Client{}
var pendingActionQueue = make(map[*Client][]*Action)
var twilioAPIKey string
var upgrader = websocket.Upgrader{}
