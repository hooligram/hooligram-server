package utils

import (
	"github.com/hooligram/logger"
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

// LogBody .
func LogBody(filePath string, text string) {
	logger.Body(
		[]string{filePath},
		text,
	)
}

// LogClose .
func LogClose(sessionID, clientID, actionType string, actionPayload interface{}) {
	logger.Close(
		[]string{sessionID, clientID, actionType},
		actionPayload,
	)
}

// LogInfo .
func LogInfo(filePath string, text string) {
	logger.Info(
		[]string{filePath},
		text,
	)
}

// LogOpen .
func LogOpen(sessionID, clientID, actionType string, actionPayload interface{}) {
	logger.Open(
		[]string{sessionID, clientID, actionType},
		actionPayload,
	)
}
