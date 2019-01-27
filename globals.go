package main

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]Client)
var httpClient = &http.Client{}
var twilioAPIKey string
var upgrader = websocket.Upgrader{}
