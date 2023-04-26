package ticketinterfaces

import (
	"time"

	"github.com/labstack/echo/v4"
)

// TicketService is an interface for managing tickets.
type BaseTicketService interface {
	Init() error
	CreateTicket(ticket Ticket) (string, error)
	UpdateTicket(ticket Ticket) error
	CloseTicket(issueKey string) error
	GetTicket(issueKey string) (Ticket, error)
	HandleWebhookAction(echo.Context) error
}

// Ticket represents a support ticket.
type Ticket struct {
	IssueKey        string    `json:"issueKey"`
	TargetContact	string	  `json:"targetContact"`
	CreationDate    time.Time `json:"creationDate"`
	Status          string    `json:"status"`
	TargetResource  string    `json:"targetResource"`
	RecommenderIDs  []string  `json:"recommenderIds"`
	LastUpdateDate time.Time `json:"lastUpdateDate"`
	LastPingDate    time.Time `json:"lastPingDate"`
	SnoozeDate      time.Time `json:"snoozeDate"`
	Subject         string    `json:"subject"`
	Assignee        string    `json:"assignee"`
}
