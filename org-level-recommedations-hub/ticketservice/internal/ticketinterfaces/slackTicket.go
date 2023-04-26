package ticketinterfaces

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/slack-go/slack"
)

type SlackTicketService struct {
	slackClient *slack.Client
	channelAsTicket bool
}

func (s *SlackTicketService) Init() error {
	apiToken := os.Getenv("SLACK_API_TOKEN")
	if apiToken == "" {
		fmt.Println("SLACK_API_TOKEN environment variable not set")
		os.Exit(1)
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
			fmt.Printf("Error parsing SLACK_CHANNEL_AS_TICKET as bool: %v\n", err)
		}
	}
	s.channelAsTicket = defaultValue
	fmt.Println("CHANNEL_AS_TICKET is set to "+strconv.FormatBool(s.channelAsTicket))
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
			fmt.Println("Channel "+channel.Name+" already exists")
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

func (s *SlackTicketService) createChannelAsTicket(ticket Ticket) (string, error) {

	channelName := fmt.Sprintf("rec-%s-%s",ticket.TargetContact,ticket.Subject)
	channelName = strings.ReplaceAll(channelName, " ", "")
	// According to this document the string length can be a max of 80
	// https://api.slack.com/methods/conversations.create
	sliceLength := 80
	stringLength := len(channelName) - 1
	if stringLength  < sliceLength {
		sliceLength = stringLength
	}
	channelName = fmt.Sprintf("%s", 
		strings.ToLower(
			channelName[0:sliceLength]))
	fmt.Println("Creating Channel: "+channelName)
	channel, err := s.createNewChannel(channelName)
	if err != nil {
		fmt.Println("Error creating channel")
		return "", err
	}

	ticket.IssueKey = channel.ID
	// Invite users to the channel (Still need to configure how users are pulled)
	userIDs := []string{ticket.Assignee}
	_, err = s.slackClient.InviteUsersToConversation(channel.ID, userIDs...)
	if err != nil {
		// If user is already in channel we should continue
		if err.Error() != "already_in_channel" {
			fmt.Println("Failed to invite users to channel:")
			return channel.ID, err
		}
		fmt.Println("User(s) were already in channel")
	}

	// Ping Channel with details of the Recommendation
	s.UpdateTicket(ticket)
	fmt.Println("Created Channel: "+channelName+"   with ID: "+channel.ID)
	return channel.ID, nil
}

func (s *SlackTicketService) createThreadAsTicket(ticket Ticket) (string, error) {
	// TODO MODIFY SO IT CREATES THREAD
	channelName := fmt.Sprintf("rec-%s-%s",ticket.TargetContact,ticket.Subject)
	// According to this document the string length can be a max of 80
	// https://api.slack.com/methods/conversations.create
	sliceLength := 80
	stringLength := len(channelName) - 1
	if stringLength  < sliceLength {
		sliceLength = stringLength
	}
	channelName = fmt.Sprintf("%s", 
		strings.ToLower(
			strings.ReplaceAll(channelName, " ", "")[0:sliceLength]))
	
	channel, err := s.createNewChannel(channelName)
	if err != nil {
		return "", err
	}

	ticket.IssueKey = channel.ID
	// Invite users to the channel (Still need to configure how users are pulled)
	userIDs := []string{ticket.Assignee}
	_, err = s.slackClient.InviteUsersToConversation(channel.ID, userIDs...)
	if err != nil {
		fmt.Println("Failed to invite users to channel: %v", err)
		return channel.ID, err
	}

	// Ping Channel with details of the Recommendation
	s.UpdateTicket(ticket)
	fmt.Println("Created Channel: "+channelName+"   with ID: "+channel.ID)
	return channel.ID, nil
}

func (s *SlackTicketService) CreateTicket(ticket Ticket) (string, error) {
	// One could argue that we should set the function on startup
	// Would save an IF statement. But meh for now
	if s.channelAsTicket{
		return s.createChannelAsTicket(ticket)
	}else {
		return s.createThreadAsTicket(ticket)
	}
}

func (s *SlackTicketService) UpdateTicket(ticket Ticket) error {
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
func (s *SlackTicketService) CloseTicket(IssueKey string) error {
	// Use the ArchiveConversation method provided by the Slack API to close the channel with the given IssueKey.
	err := s.slackClient.ArchiveConversation(IssueKey)
	if err != nil {
		// If there's an error while closing the channel, return the error.
		return err
	}
	// If the channel was successfully closed, return nil.
	return nil
}

// Incomplete
func (s *SlackTicketService) GetTicket(issueKey string) (Ticket, error) {
	conversationInfo, err := s.slackClient.GetConversationInfo(
		&slack.GetConversationInfoInput{
			ChannelID:     issueKey,
			IncludeLocale: false,
		})
	if err != nil {
		return Ticket{}, err
	}
	ticket := Ticket{
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
	fmt.Println("Received message: %s\n", action)

	// Return nil to indicate that there was no error
	return nil
}
