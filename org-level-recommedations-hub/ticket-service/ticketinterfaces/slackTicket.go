package ticketinterfaces

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/slack-go/slack"
)

type SlackTicketService struct {
	slackClient slack.Client
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
	return nil
}

func (s *SlackTicketService) createChannel( issueKey int) (string, error) {
	channelName := fmt.Sprintf("rec-%d", issueKey)

	// Check if channel already exists
	channels, err := s.slackClient.GetChannels(false)
	if err != nil {
		return "", err
	}
	for _, channel := range channels {
		if channel.Name == channelName {
			return channel.ID, nil
		}
	}

	// Create channel if it doesn't exist
	channel, err := s.slackClient.CreateChannel(channelName)
	if err != nil {
		return "", err
	}
	// Invite users to the channel (Still need to configure how users are pulled)
	userIDs := []string{"USER_ID_1", "USER_ID_2"}
	for _, userID := range userIDs {
		err := s.slackClient.InviteUsersToConversation(channel.ID, userID)
		if err != nil {
			fmt.Printf("Failed to invite user %s to channel: %v", userID, err)
			return
		}
	}

	return channel.ID, nil
}

func (s *SlackTicketService) CreateTicket(ticket Ticket) (string,error) {
	// Still need to set channel.ID to IssueKey, that's gonna be one of the problems here I need to sort out
	return s.createChannel(ticket.IssueKey)
}

func (s *SlackTicketService) UpdateTicket(ticket Ticket) error {
	_, _, err := s.slackClient.PostMessage(
        ticket.IssueKey,
        slack.MsgOptionText(string(json.Marshal(ticket)), false),
    )
    if err != nil {
        return err
    }
    return nil
}

// CloseTicket is a function that closes an existing channel in Slack based on the given IssueKey.
func (s *SlackTicketService) CloseTicket(IssueKey string) error {
	// Use the ArchiveConversation method provided by the Slack API to close the channel with the given IssueKey.
	_, err := s.slackClient.ArchiveConversation(IssueKey)
	if err != nil {
		// If there's an error while closing the channel, return the error.
		return err
	}
	// If the channel was successfully closed, return nil.
	return nil
}

// Incomplete
func (s *SlackTicketService) GetTicket(IssueKey string) (Ticket, error) {
	ticket := Ticket{
		IssueKey: "123"
	}
	return Ticket
}

type Message struct {
    Token       string      `json:"token"`
    TeamID      string      `json:"team_id"`
    APIAppID    string      `json:"api_app_id"`
    Event       Event       `json:"event"`
    Text        string      `json:"text"`
    Type        string      `json:"type"`
    AuthedUsers []string    `json:"authed_users"`
}

type Event struct {
    Type            string          `json:"type"`
    User            string          `json:"user"`
    Text            string          `json:"text"`
    Ts              string          `json:"ts"`
    Channel         string          `json:"channel"`
    EventTimestamp  json.RawMessage `json:"event_ts"`
}

func (s *SlackTicketService) HandleWebhookAction(w http.ResponseWriter, r *http.Request) error {
	// Check if the HTTP method is POST
	if r.Method != "POST" {
		// If the method is not POST, return a 405 Method Not Allowed error
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}

	// Decode the request body into a Message struct
	decoder := json.NewDecoder(r.Body)
	var msg Message
	err := decoder.Decode(&msg)
	if err != nil {
		// If there is an error decoding the request body, return a 400 Bad Request error
		http.Error(w, "Bad request", http.StatusBadRequest)
	}

	// Print the received message to the console
	fmt.Printf("Received message: %s\n", msg.Text)

	// Return nil to indicate that there was no error
	return nil
}