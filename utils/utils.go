package utils

import (
	"math/rand"

	"github.com/hooligram/kifu"
)

// ContainsID .
func ContainsID(ids []int, id int) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}

	return false
}

// GenerateRandomString .
func GenerateRandomString(stringLength int) string {
	if stringLength < 1 {
		return ""
	}

	var runes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	randomRunes := make([]rune, stringLength)

	for i := range randomRunes {
		randomRunes[i] = runes[rand.Intn(len(runes))]
	}

	return string(randomRunes)
}

// LogBody .
func LogBody(filePath string, text string) {
	kifu.Body(
		[]string{filePath},
		text,
	)
}

// LogClose .
func LogClose(sessionID, clientID, actionType string, actionPayload interface{}) {
	kifu.Close(
		[]string{sessionID, clientID, actionType},
		actionPayload,
	)
}

// LogFatal .
func LogFatal(filePath string, text string) {
	kifu.Fatal(
		[]string{filePath},
		text,
	)
}

// LogInfo .
func LogInfo(filePath string, text string) {
	kifu.Info(
		[]string{filePath},
		text,
	)
}

// LogOpen .
func LogOpen(sessionID, clientID, actionType string, actionPayload interface{}) {
	kifu.Open(
		[]string{sessionID, clientID, actionType},
		actionPayload,
	)
}
