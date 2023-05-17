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
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"regexp"
	"log"
	"net/http"
	"os"
	"time"
	"strconv"
	"strings"
	u "ticketservice/internal/utils"
	t "ticketservice/internal/ticketinterfaces"
	b "ticketservice/internal/bigqueryfunctions"
	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

type Command struct {
	Token       string `json:"token"`
	Command     string `json:"command"`
	Text        string `json:"text"`
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	ResponseURL string `json:"response_url"`
}


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

func (s *SlackTicketService) createChannelAsTicket(ticket *t.Ticket, row t.RecommendationQueryResult) (string, error) {
	lastSlashIndex := strings.LastIndex(row.TargetResource, "/")
	secondToLast := strings.LastIndex(row.TargetResource[:lastSlashIndex], "/")
	// This could be moved to BQ Query. But ehh
	ticket.CreationDate = time.Now()
	ticket.LastUpdateDate = time.Now()
	ticket.LastPingDate = time.Now()
	ticket.SnoozeDate = time.Now().AddDate(0,0,7)
	ticket.Subject = fmt.Sprintf("%s-%s",
			row.Recommender_subtype,
			nonAlphanumericRegex.ReplaceAllString(
				row.TargetResource[secondToLast+1:],
				""))
	ticket.RecommenderID = row.Recommender_name
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
	s.UpdateTicket(ticket, row)
	u.LogPrint(2,"Created Channel: "+channelName+"   with ID: "+channel.ID)
	return channel.ID, nil
}

func (s *SlackTicketService) createThreadAsTicket(ticket *t.Ticket, row t.RecommendationQueryResult) (string, error) {
	lastSlashIndex := strings.LastIndex(row.TargetResource, "/")
	secondToLast := strings.LastIndex(row.TargetResource[:lastSlashIndex], "/")
	// This could be moved to BQ Query. But ehh
	ticket.CreationDate = time.Now()
	ticket.LastUpdateDate = time.Now()
	ticket.LastPingDate = time.Now()
	ticket.SnoozeDate = time.Now().AddDate(0,0,7)
	ticket.Subject = fmt.Sprintf("%s-%s-%s",
			row.Project_name,
			nonAlphanumericRegex.ReplaceAllString(
				row.TargetResource[secondToLast+1:],
				""),
			row.Recommender_subtype)
	ticket.RecommenderID = row.Recommender_name
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
	messageOptions := slack.MsgOptionText(ticket.Subject, false)
	_ ,timestamp, err := s.slackClient.PostMessage(channel.ID, messageOptions)
	if err != nil {
		u.LogPrint(3, "Failed to send message to channel")
		return channel.ID, err
	}

	ticket.IssueKey = channel.ID + "-" + timestamp

	s.UpdateTicket(ticket, row)
	u.LogPrint(1, "Created Ticket in Channel: "+channelName+" with ID: "+timestamp)
	return ticket.IssueKey, nil
}

func (s *SlackTicketService) CreateTicket(ticket *t.Ticket, row t.RecommendationQueryResult) (string, error) {
	// One could argue that we should set the function on startup
	// Would save an IF statement. But meh for now
	if s.channelAsTicket{
		return s.createChannelAsTicket(ticket, row)
	}else {
		return s.createThreadAsTicket(ticket, row)
	}
}

// TODO (Ghaun): Update this to take in channel as ticket vs not.
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

// Incomplete
func (s *SlackTicketService) GetTicket(issueKey string) (t.Ticket, error) {
	// Slack tickets are super simple, so let's pull from BQ
	ticket, err := b.GetTicketByIssueKey(issueKey)
	if err != nil {
		// Handle the error
	}
	return *ticket, nil
}

func sendFastResponseAndProcess(body []byte, c echo.Context) error {
	// Get the URL from the request
	url := fmt.Sprintf("https://%s%s", c.Request().Host, c.Request().URL.Path)
	fmt.Printf("Resend to: %s\n", url)

	// Create a new request with the received URL
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}

	// Forward the request with an additional header
	req.Header.Set("X-PROCESS-SLACK", "true")

	// Add the X-Slack-Signature header
	signature := c.Request().Header.Get("X-Slack-Signature")
	timestamp := c.Request().Header.Get("X-Slack-Request-Timestamp")
	req.Header.Set("X-Slack-Signature", signature)
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)

	// Create a custom RoundTripper to ignore the response body
	transport := &http.Transport{
		DisableKeepAlives: true,
	}
	client := &http.Client{Transport: transport}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	go func() {
	// Perform the request, waiting for the request to finish sending
		_, err = client.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
		}
	}()
	// Small Delay to ensure the request is completely sent.
	time.Sleep(time.Second)
	c.NoContent(http.StatusAccepted)
	return nil
}

func (s *SlackTicketService) HandleWebhookAction(c echo.Context) error {
	// So the problem with Slack is that they expect a response within 3 seconds
	// https://api.slack.com/apis/connections/events-api#responding
	// So we actually need to send a response quickly and process in the background.
	// BQ as our main datasource means it's impossible to respond within 3 seconds.

	// We MUST respond within 3 seconds or a retry happens, however we could solve
	// The retry issue with some form of caching, however that doesn't completely solve
	// the problem, because Slack will disable events if the majority of events are over 3s

	// Here is where it get's interesting, we want to support serverless deployment
	// Serverless services generally shut down processing once a response has been sent.
	// There are many ways to solve this problem, but in most cases it uses another outside service
	// Ideally this should be able to run on any platform, regardless of deployment.

	// I'm including it only in the Slack Plugin, because this may not be a problem for other plugins
	// Happy to move it if we need.

	// So to get around this, and NOT use other services (PubSub, Cloud Tasks) we will call
	// the Webhook again, but with a header that allows the service to actually process.

	//Verifying the request is ALWAYS first. 
	defer c.Request().Body.Close()
	body, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return err
	}
	if !verifyRequestSignature(c.Request().Header, body) {
		return fmt.Errorf("Failed to Verify Request Signature")
	}

	// Check if the X-PROCESS-SLACK header is set to true
	processSlack := c.Request().Header.Get("X-PROCESS-SLACK") == "true"
	if !processSlack {
		u.LogPrint(1, "Recieved first webhook, will process in background")
		sendFastResponseAndProcess(body, c)
		return c.NoContent(http.StatusOK)
	}

    // Parse the event payload
    u.LogPrint(1, "Body: %v", string(body))
    var event slackevents.EventsAPICallbackEvent
    err = json.Unmarshal(body, &event)
    if err != nil {
        return err
    }
    switch event.Type {
    case slackevents.URLVerification:
        var r *slackevents.ChallengeResponse
        err := json.Unmarshal(body, &r)
        if err != nil {
            return err
        }
        return c.JSON(http.StatusOK, r)

    case slackevents.CallbackEvent:
		var eventType *slackevents.EventsAPIInnerEvent
		err = json.Unmarshal([]byte(*event.InnerEvent), &eventType)
		if err != nil {
            return err
        }
        if slackevents.EventsAPIType(eventType.Type) == slackevents.Message {
            // Unmarshal the inner event into a MessageEvent
            var messageEvent *slackevents.MessageEvent
            err = json.Unmarshal(*event.InnerEvent, &messageEvent)
            if err != nil {
                return err
            }
            // Now you have access to the message event data
            u.LogPrint(1, "Received message event: %v", messageEvent)
			messageSplitBySpaces := strings.Split(messageEvent.Text, " ")
			if len(messageSplitBySpaces) < 1 {
				u.LogPrint(1, "Message did not have any length")
				return nil
			}
			// Check if it's a command
			command := strings.ToLower(messageSplitBySpaces[0])
			if !regexp.MustCompile(`^!`).MatchString(command) {
				u.LogPrint(1, "Not a command: %v", command)
				return nil
			}
			// Now let's fire up the correct command
			if function, ok := functionMap[command]; ok {
				// Call the function with the event and arguments
				err := function(s, messageEvent, messageSplitBySpaces)
				if err != nil {
					u.LogPrint(3, "Something went wrong with function: %v", command)
				}
				u.LogPrint(2, "Completed Function %v", command)
			} else {
				// Command not found
				u.LogPrint(1, "Command %v not found", command)
				return nil
			}
			return nil

        }

    default:
        return c.String(http.StatusInternalServerError, fmt.Sprintf("Unexpected event type: %s", event.Type))
    }
    return nil
}

func verifyRequestSignature(header http.Header, body []byte) bool {
    // Extract the signature and timestamp from the header
    signature := header.Get("X-Slack-Signature")
    timestamp := header.Get("X-Slack-Request-Timestamp")
    // Ensure the timestamp is not too old
    timestampInt, err := strconv.Atoi(timestamp)
    if err != nil {
		u.LogPrint(2, "Verify Request Signature Failed at strconv.Atoi: %v", err)
        return false
    }
    age := time.Now().Unix() - int64(timestampInt)
    if age > 300 {
        return false
    }

    // Concatenate the timestamp and request body
    sigBaseString := fmt.Sprintf("v0:%s:%s", timestamp, string(body))
    // Hash the base string with the Slack signing secret
    signatureHash := hmac.New(sha256.New, []byte(slackSigningSecret))
    signatureHash.Write([]byte(sigBaseString))
    expectedSignature := fmt.Sprintf("v0=%s", hex.EncodeToString(signatureHash.Sum(nil)))

    // Compare the expected signature to the actual signature
	equal := hmac.Equal([]byte(signature), []byte(expectedSignature))
	u.LogPrint(1, "Received Sig: %s   Calculated Sig: %s", signature, expectedSignature)
    return equal
}

// C = Channel, t = ThreadTimeStamp, m = message you want to send
func (s *SlackTicketService) sendSlackMessage(c string, t string, m string) error{
	// Send the message to the channel in which the event occurred
	u.LogPrint(1, "Sending message to channel: %s, timestamp: %s, with message: %s", c,t,m)
	message := slack.MsgOptionText(m, false)
	if !s.channelAsTicket {
		_, _, _, err := s.slackClient.SendMessage(c, slack.MsgOptionTS(t), message)
		if err != nil {
			u.LogPrint(3, "Failed to respond in thread: %v", err)
			return err
		}
		return nil
	}
	_, _, err := s.slackClient.PostMessage(c, message)
	if err != nil {
		u.LogPrint(3,"Error sending message: %s\n", err)
		return err
	}
	return nil
}

func (s *SlackTicketService) parseAndGetTicket(channel, timestamp string) (t.Ticket, error) {
	issueKey := channel
	if !s.channelAsTicket {
		issueKey = fmt.Sprintf("%v-%v", channel, timestamp)
	}
	ticket, err := b.GetTicketByIssueKey(issueKey)
	if err != nil {
		u.LogPrint(3, "[SLACK] Error getting ticket from Bigquery: %v", err)
		return t.Ticket{}, err
	}
	return *ticket, nil
}

// This code should probably be moved to a different file....
// But oh well for now. It would require me to modify the compile script
// And I'd rather get this working first.
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
