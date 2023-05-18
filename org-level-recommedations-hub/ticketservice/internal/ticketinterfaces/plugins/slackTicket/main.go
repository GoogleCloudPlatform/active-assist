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
	"log"
	"os"
	"regexp"
	"strconv"

	"github.com/slack-go/slack"
	
	t "ticketservice/internal/ticketinterfaces"
	u "ticketservice/internal/utils"
)

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)
var slackSigningSecret = ""

type SlackTicketService struct {
	slackClient *slack.Client
	slackSigningSecret	string
	channelAsTicket bool
}

func CreateService() t.BaseTicketService{
	var service SlackTicketService
	return &service
}

func (s *SlackTicketService) Init() error {
	apiToken := os.Getenv("SLACK_API_TOKEN")
	if apiToken == "" {
		u.LogPrint(4,"SLACK_API_TOKEN environment variable not set")
	}
	ss := os.Getenv("SLACK_SIGNING_SECRET")
	if ss == "" {
		u.LogPrint(4,"SLACK_SIGNING_SECRET environment variable not set")
	}
	slackSigningSecret = ss
	// Create a new Slack client with your API token
	s.slackClient = slack.New(apiToken)

	// Use the Slack client in your code
	_, err := s.slackClient.AuthTest()
	if err != nil {
		log.Fatalf("Error authenticating with Slack: %s", err)
	}
	log.Println("Successfully authenticated with Slack!")
	// Let's see if the environment wants to use channel as ticket
	// or thread as ticket
	cAsT := os.Getenv("SLACK_CHANNEL_AS_TICKET")
	defaultValue := true
	if cAsT != "" {
		var err error
		defaultValue, err = strconv.ParseBool(cAsT)
		if err != nil {
			u.LogPrint(3,"Error parsing SLACK_CHANNEL_AS_TICKET as bool: %v\n", err)
		}
	}
	s.channelAsTicket = defaultValue
	u.LogPrint(1,"CHANNEL_AS_TICKET is set to "+strconv.FormatBool(s.channelAsTicket))
	return nil
}