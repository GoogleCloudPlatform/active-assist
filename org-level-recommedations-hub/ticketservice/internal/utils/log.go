// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.```

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
		log.Printf("Unable to parse log string: %v", v)
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
