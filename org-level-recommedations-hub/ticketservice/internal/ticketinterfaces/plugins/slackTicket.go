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

package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"log"
	"os"
	"strconv"
	"strings"
	u "ticketservice/internal/utils"
	t "ticketservice/internal/ticketinterfaces"
	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
)

type SlackTicketService struct {
	slackClient *slack.Client
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

func (s *SlackTicketService) createNewChannel(channelName string) (*slack.Channel, error){
	// Check if channel already exists
	channels, _, err := s.slackClient.GetConversations(&slack.GetConversationsParameters{
		ExcludeArchived: true,
	})
	if err != nil {
		return nil, err
	}
	// One could argue we could store this result in memory or some form of memorystore.
	// But I'm not sure the length here would get to a performance impact. Happy to adjust
	// in the future
	for _, channel := range channels {
		if channel.Name == channelName {
			u.LogPrint(1,"Channel "+channel.Name+" already exists")
			return &channel, nil
		}
	}
	// Create channel if it doesn't exist
	channel, err := s.slackClient.CreateConversation(slack.CreateConversationParams{
		ChannelName: channelName,
	})
	if err != nil {
		return nil, err
	}
	return channel, nil
}

func (s *SlackTicketService) createChannelAsTicket(ticket t.Ticket) (string, error) {

	channelName := fmt.Sprintf("rec-%s-%s",ticket.TargetContact,ticket.Subject)
	channelName = strings.ReplaceAll(channelName, " ", "")
	// According to this document the string length can be a max of 80
	// https://api.slack.com/methods/conversations.create
	sliceLength := 80
	stringLength := len(channelName) - 1
	if stringLength  < sliceLength {
		sliceLength = stringLength
	}
	channelName = strings.ToLower(channelName[0:sliceLength])
	u.LogPrint(1,"Creating Channel: "+channelName)
	channel, err := s.createNewChannel(channelName)
	if err != nil {
		u.LogPrint(3,"Error creating channel")
		return "", err
	}

	ticket.IssueKey = channel.ID
	_, err = s.slackClient.InviteUsersToConversation(channel.ID, ticket.Assignee...)
	if err != nil {
		// If user is already in channel we should continue
		if err.Error() != "already_in_channel" {
			u.LogPrint(3,"Failed to invite users to channel:")
			return channel.ID, err
		}
		u.LogPrint(1,"User(s) were already in channel")
	}

	// Ping Channel with details of the Recommendation
	s.UpdateTicket(ticket)
	u.LogPrint(1,"Created Channel: "+channelName+"   with ID: "+channel.ID)
	return channel.ID, nil
}

func (s *SlackTicketService) createThreadAsTicket(ticket t.Ticket) (string, error) {
	channelName := strings.ToLower(ticket.TargetContact)

	// Replace multiple characters using regex to conform to Slack channel name restrictions
	re := regexp.MustCompile(`[\s@#._/:\\*?"<>|]+`)
	channelName = re.ReplaceAllString(channelName, "-")

	u.LogPrint(1, "Creating Channel: "+channelName)
	channel, err := s.createNewChannel(channelName)
	if err != nil {
		u.LogPrint(3, "Error creating channel")
		return "", err
	}

	// Invite users to the channel
	_, err = s.slackClient.InviteUsersToConversation(channel.ID, ticket.Assignee...)
	if err != nil {
		// If user is already in channel we should continue
		if err.Error() != "already_in_channel" {
			u.LogPrint(3,"Failed to invite users to channel:")
			return channel.ID, err
		}
		u.LogPrint(1,"User(s) were already in channel")
	}

	// Send message to the created channel to create "ticket/thread"
	messageContent := ticket.Subject
	messageOptions := slack.MsgOptionText(messageContent, false)
	_ ,timestamp, err := s.slackClient.PostMessage(channel.ID, messageOptions)
	if err != nil {
		u.LogPrint(3, "Failed to send message to channel")
		return channel.ID, err
	}

	// Respond in thread with the JSON representation of the ticket
	jsonData, err := json.Marshal(ticket)
	if err != nil {
		u.LogPrint(3, "Failed to marshal ticket to JSON")
		return channel.ID, err
	}

	threadMessageOptions := slack.MsgOptionText(string(jsonData), false)
	_, _, _, err = s.slackClient.SendMessage(channel.ID, slack.MsgOptionTS(timestamp), threadMessageOptions)
	if err != nil {
		u.LogPrint(3, "Failed to respond in thread")
		return channel.ID, err
	}

	ticket.IssueKey = channelName + "-" + timestamp

	s.UpdateTicket(ticket)
	u.LogPrint(1, "Created Ticket in Channel: "+channelName+" with ID: "+timestamp)
	return ticket.IssueKey, nil
}

func (s *SlackTicketService) CreateTicket(ticket t.Ticket) (string, error) {
	// One could argue that we should set the function on startup
	// Would save an IF statement. But meh for now
	if s.channelAsTicket{
		return s.createChannelAsTicket(ticket)
	}else {
		return s.createThreadAsTicket(ticket)
	}
}

// TODO (Ghaun): Update this to take in channel as ticket vs not.
func (s *SlackTicketService) UpdateTicket(ticket t.Ticket) error {
	jsonData, err := json.MarshalIndent(ticket, "", "    ")
	if err != nil {
		return err
	}
	//Convert to code block
	message := fmt.Sprintf("```%s```", string(jsonData))
	_, _, err = s.slackClient.PostMessage(
		ticket.IssueKey,
		slack.MsgOptionText(message, false),
	)
	if err != nil {
		return err
	}
	return nil
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

// Incomplete
func (s *SlackTicketService) GetTicket(issueKey string) (t.Ticket, error) {
	conversationInfo, err := s.slackClient.GetConversationInfo(
		&slack.GetConversationInfoInput{
			ChannelID:     issueKey,
			IncludeLocale: false,
		})
	if err != nil {
		return t.Ticket{}, err
	}
	ticket := t.Ticket{
		IssueKey: conversationInfo.ID,
		// Need to determinet the best way to get the ticket information back from slack
		// Will need to do this once testing begings
	}
	return ticket, nil
}

type Message struct {
	Token       string   `json:"token"`
	TeamID      string   `json:"team_id"`
	APIAppID    string   `json:"api_app_id"`
	Event       Event    `json:"event"`
	Text        string   `json:"text"`
	Type        string   `json:"type"`
	AuthedUsers []string `json:"authed_users"`
}

type Event struct {
	Type           string          `json:"type"`
	User           string          `json:"user"`
	Text           string          `json:"text"`
	Ts             string          `json:"ts"`
	Channel        string          `json:"channel"`
	EventTimestamp json.RawMessage `json:"event_ts"`
}

// Haven't determined what all this will do yet.
func (s *SlackTicketService) HandleWebhookAction(c echo.Context) error {

	// Decode the request body into a Message struct
	action := c.Param("action")

	// Print the received message to the console
	u.LogPrint(1,"Received message: %s\n", action)

	// Return nil to indicate that there was no error
	return nil
}
