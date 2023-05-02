package utils

import (
	"log"
	"os"
	"strconv"
)

const (
	Debug = 1
	Info = 2
	Warning = 3
	Error = 4
)

var logLevel int

func LogPrint(level int, v ...interface{}) {
	// Check if the initial log level variable is set
	if logLevel == 0 {
		// Read log level from environment variable
		envLevel, err := strconv.Atoi(os.Getenv("LOG_LEVEL"))
		if err != nil || envLevel < Debug || envLevel > Error {
			// Default log level if environment variable is invalid or not set
			logLevel = Error
		} else {
			logLevel = envLevel
		}
	}

	// Check if the first argument is a string
	f, ok := v[0].(string)
	if !ok {
		log.Print("Unable to parse log string")
	}
	// Remove the first value of the array
	v = v[1:]

	if level == Error {
		// Fatal log level - log and exit
		log.Fatalf(f, v...)
	}

	if level >= logLevel {
		// Log message if log level is equal or higher than the current log level
		log.Printf(f, v...)
	}
}
