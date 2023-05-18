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
// limitations under the License.

package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack/slackevents"

	b "ticketservice/internal/bigqueryfunctions"
	u "ticketservice/internal/utils"
)

var functionMap = map[string]func(*SlackTicketService, *slackevents.MessageEvent, []string) error{
	// All commands should be lower case.
	"!snooze": snoozeFunction,
}

func snoozeFunction(s *SlackTicketService, event *slackevents.MessageEvent, splitText []string) error {
	if len(splitText) < 2 {
		u.LogPrint(1, "Did not recieve enough arguments for Snooze. IE. !Snooze for x days")
		// Send a message response here.
		return s.sendSlackMessage(event.Channel, event.ThreadTimeStamp, "Not enough arguments")
	}
	
	response := strings.Join(splitText[1:], " ")

	// Define regular expressions to match the numeric value and the duration unit
	valueRegex := regexp.MustCompile(`(\d+)`)
	unitRegex := regexp.MustCompile(`\b(days?|months?|years?)\b`)

	// Extract the numeric value and the duration unit from the response
	valueMatches := valueRegex.FindStringSubmatch(response)
	unitMatches := unitRegex.FindStringSubmatch(response)

	if len(valueMatches) < 2 || len(unitMatches) < 2 {
		u.LogPrint(2, "Failed to extract duration from the response")
		// Send a message response here.
		return s.sendSlackMessage(event.Channel, event.ThreadTimeStamp, "Invalid duration format")
	}

	// Parse the numeric value from the matches
	value, err := strconv.Atoi(valueMatches[1])
	if err != nil {
		u.LogPrint(2, "Failed to parse duration value:", err)
		// Send a message response here.
		return s.sendSlackMessage(event.Channel, event.ThreadTimeStamp, "Invalid duration format")
	}

	// Convert the duration unit to lowercase for consistency
	unit := strings.ToLower(unitMatches[1])

	// Map the duration unit to the corresponding time unit
	var timeUnit time.Duration
	switch unit {
	case "day", "days":
		timeUnit = time.Hour * 24
	case "month", "months":
		timeUnit = time.Hour * 24 * 30
	case "year", "years":
		timeUnit = time.Hour * 24 * 365
	default:
		u.LogPrint(2, "Invalid duration unit")
		// Send a message response here.
		return s.sendSlackMessage(event.Channel, event.ThreadTimeStamp, "Invalid duration unit")
	}

	// Calculate the total duration based on the value and the time unit
	duration := time.Duration(value) * timeUnit

	// Now you have the parsed duration as a time.Duration object
	u.LogPrint(2, "Parsed duration:", duration)
	ticket, err := s.parseAndGetTicket(event.Channel, event.ThreadTimeStamp)
	if err != nil {
		return s.sendSlackMessage(event.Channel, event.ThreadTimeStamp, "Something went wrong getting ticket")
	}
	ticket.SnoozeDate = time.Now().Add(duration)
	if err := b.UpsertTicket("", ticket); err != nil {
		u.LogPrint(3, "[SLACK] Something went wrong updating ticket in BQ: %v", err)
		return s.sendSlackMessage(event.Channel, event.ThreadTimeStamp, "Something went wrong")
	}
	
	return s.sendSlackMessage(event.Channel, event.ThreadTimeStamp, fmt.Sprintf("Snoozed Until: %d", ticket.SnoozeDate))
}
