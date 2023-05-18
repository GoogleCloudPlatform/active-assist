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
	"encoding/json"
	"fmt"
	"strings"

	b "ticketservice/internal/bigqueryfunctions"
	t "ticketservice/internal/ticketinterfaces"
)


func (s *SlackTicketService) CreateTicket(ticket *t.Ticket, row t.RecommendationQueryResult) (string, error) {
	// One could argue that we should set the function on startup
	// Would save an IF statement. But meh for now
	if s.channelAsTicket{
		return s.createChannelAsTicket(ticket, row)
	}else {
		return s.createThreadAsTicket(ticket, row)
	}
}

func (s *SlackTicketService) UpdateTicket(ticket *t.Ticket, row t.RecommendationQueryResult) error {
	jsonData, err := json.MarshalIndent(ticket, "", "    ")
	if err != nil {
		return err
	}
	//Convert to code block
	message := fmt.Sprintf("```%s```\n Cost Savings:%v in %v \n Description: %v", 
		string(jsonData),
		row.Impact_cost_unit,
		row.Impact_currency_code,
		row.Description)
	if !s.channelAsTicket {
		// This will return an array. [0] will be channel id [1] will be timestamp
		channelTimestamp := strings.Split(ticket.IssueKey, "-")
		return s.sendSlackMessage(channelTimestamp[0], channelTimestamp[1], message)
	}
	return s.sendSlackMessage(ticket.IssueKey, "", message)
}

// CloseTicket is a function that closes an existing channel in Slack based on the given IssueKey.
func (s *SlackTicketService) CloseTicket(key string) error {
	// Use the ArchiveConversation method provided by the Slack API to close the channel with the given IssueKey.
	err := s.slackClient.ArchiveConversation(key)
	if err != nil {
		// If there's an error while closing the channel, return the error.
		return err
	}
	// If the channel was successfully closed, return nil.
	return nil
}

func (s *SlackTicketService) GetTicket(issueKey string) (t.Ticket, error) {
	// Slack tickets are super simple, so let's pull from BQ
	ticket, err := b.GetTicketByIssueKey(issueKey)
	if err != nil {
		// Handle the error
	}
	return *ticket, nil
}