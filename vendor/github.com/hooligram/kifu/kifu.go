package kifu

import (
	"fmt"
	"log"
)

// Tags to more easily identify log types.
const (
	BODY  = "BODY"
	CLOSE = "CLSE"
	FATAL = "FATL"
	INFO  = "INFO"
	OPEN  = "OPEN"
)

// Body .
func Body(identifiers []string, content interface{}) {
	logText := constructLogText(BODY, identifiers, content)
	log.Println(logText)
}

// Close .
func Close(identifiers []string, content interface{}) {
	logText := constructLogText(CLOSE, identifiers, content)
	log.Println(logText)
}

// Fatal .
func Fatal(identifiers []string, content interface{}) {
	logText := constructLogText(FATAL, identifiers, content)
	log.Fatalln(logText)
}

// Info .
func Info(identifiers []string, content interface{}) {
	logText := constructLogText(INFO, identifiers, content)
	log.Println(logText)
}

// Open .
func Open(identifiers []string, content interface{}) {
	logText := constructLogText(OPEN, identifiers, content)
	log.Println(logText)
}

func constructLogText(logType string, identifiers []string, content interface{}) string {
	logText := logType

	for _, identifier := range identifiers {
		logText += fmt.Sprintf(" [%v]", identifier)
	}

	if content != nil {
		logText += fmt.Sprintf(" %v", content)
	}

	return logText
}
