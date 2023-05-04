``` Copyright 2020 Google LLC

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.```
 
package ticketinterfaces

import (
	"time"
	"plugin"
	"log"
	"github.com/labstack/echo/v4"
)

// Your plugin needs to have the method CreateService that returns your BaseTicketService interface implementation

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
	RecommenderID  string  `json:"recommenderIds"`
	LastUpdateDate time.Time `json:"lastUpdateDate"`
	LastPingDate    time.Time `json:"lastPingDate"`
	SnoozeDate      time.Time `json:"snoozeDate"`
	Subject         string    `json:"subject"`
	Assignee        []string    `json:"assignee"`
}

func InitTicketService(implName string) (BaseTicketService, error) {

	// Load the plugin based on the name
	pluginPath := "./plugins/" + implName + ".so"
	p, err := plugin.Open(pluginPath)
	if err != nil {
		log.Fatalf("Failed to open plugin: %v", err)
	}

	// Look up the "NewTicketService" symbol in the plugin
	newTicketServiceSymbol, err := p.Lookup("CreateService")
	if err != nil {
		log.Fatalf("Failed to lookup symbol: %v", err)
	}

	// Create an instance of the ticket service implementation
	implValue := newTicketServiceSymbol.(func() BaseTicketService)()

	// Initialize the ticket service implementation
	if err := implValue.Init(); err != nil {
		return nil, err
	}

	return implValue, nil
}