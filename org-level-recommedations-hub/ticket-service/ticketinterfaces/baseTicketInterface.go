package ticketinterfaces

import (
	"net/http"
	"time"
)

// TicketService is an interface for managing tickets.
type BaseTicketService interface {
	Init() error
	CreateTicket(ticket Ticket) (string, error)
	UpdateTicket(ticket Ticket) error
	CloseTicket(issueKey string) error
	GetTicket(issueKey string) (Ticket, error)
	HandleWebhookAction(w http.ResponseWriter, r *http.Request) error
}

// Ticket represents a support ticket.
type Ticket struct {
	IssueKey        string    `json:"issueKey"`
	CreationDate    time.Time `json:"creationDate"`
	Status          string    `json:"status"`
	TargetResource  string    `json:"targetResource"`
	RecommenderIDs  []string  `json:"recommenderIds"`
	LastUpdatedDate time.Time `json:"lastUpdatedDate"`
	LastPingDate    time.Time `json:"lastPingDate"`
	SnoozeDate      time.Time `json:"snoozeDate"`
	Subject         string    `json:"subject"`
	Assignee        string    `json:"assignee"`
}
