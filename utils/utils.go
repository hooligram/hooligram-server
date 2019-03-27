package utils

import (
	"math/rand"
	"strings"

	"github.com/hooligram/kifu"
)

// ContainsInt .
func ContainsInt(ints []int, target int) bool {
	for _, i := range ints {
		if i == target {
			return true
		}
	}

	return false
}

// ContainsString .
func ContainsString(strings []string, target string) bool {
	for _, s := range strings {
		if s == target {
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

// ParseSID .
func ParseSID(sid string) (string, string) {
	parsed := strings.Split(sid, ".")
	if len(parsed) < 2 {
		return "", ""
	}

	return parsed[0], parsed[1]
}
